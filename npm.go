package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
)

type NodeVersion struct {
	Version string `json:"version"`
	Npm     string `json:"npm"`
}

type PackageVersionInfo struct {
	Versions map[string]map[string]interface{} `json:"versions"`
}

type PackageInfo struct {
	Engines map[string]interface{} `json:"engines"`
}

type PackageJson struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         map[string]string `json:"engines"`
	Name            string            `json:"name"`
	Version         string            `json:"version"`
}

type MinMax bool

const (
	Min = false
	Max = true
)

var getAllNodeVersionsUrl string = "https://nodejs.org/dist/index.json"
var npmRegistryBaseUrl string = "https://registry.npmjs.com"

func getAllNodeSemvers(ltsOnly bool) ([]*semver.Version, error) {
	var nodeVersions []NodeVersion
	var nodeSemvers = make([]*semver.Version, 0, 100)
	var client http.Client
	response, err := client.Get(getAllNodeVersionsUrl)
	if err != nil {
		err = fmt.Errorf("GetAllNodeVersions: %s", color.RedString("error fetching Node versions"))
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		nodeJson, _ := io.ReadAll(response.Body)
		json.Unmarshal(nodeJson, &nodeVersions)
	}

	for _, node := range nodeVersions {
		nodeSemver, err := semver.NewVersion(strings.Replace(node.Version, "v", "", 1))
		if err == nil && (!ltsOnly || nodeSemver.Major()%2 == 0) {
			nodeSemvers = append(nodeSemvers, nodeSemver)
		}
	}

	slices.SortFunc(nodeSemvers, func(a, b *semver.Version) int {
		return a.Compare(b)
	})

	return nodeSemvers, err
}

func getPackageVersions(pkg string) ([]string, error) {
	var client http.Client
	versionsResponse, err := client.Get(fmt.Sprintf("%s/%s", npmRegistryBaseUrl, pkg))

	if err != nil {
		return make([]string, 0), err
	}

	if versionsResponse.StatusCode == http.StatusOK {
		versions, err := io.ReadAll(versionsResponse.Body)
		if err != nil {
			return make([]string, 0), err
		}
		var pkgVersionInfo PackageVersionInfo
		err = json.Unmarshal(versions, &pkgVersionInfo)
		if err != nil {
			return make([]string, 0), err
		}
		pkgVersions := make([]string, 0)

		for version := range pkgVersionInfo.Versions {
			pkgVersions = append(pkgVersions, version)
		}
		return pkgVersions, nil
	}
	return make([]string, 0), nil
}

func getSuitableVersionString(max MinMax, versionRange string, depVersions []string) string {
	versionConstraint, err := semver.NewConstraint(versionRange)
	if err != nil {
		return versionRange
	}
	candidateVersions := make([]*semver.Version, 0, 10)

	for _, depVersion := range depVersions {
		depSemver, err := semver.NewVersion(depVersion)
		if err != nil {
			continue
		}
		if versionConstraint.Check(depSemver) {
			candidateVersions = append(candidateVersions, depSemver)
		}
	}
	slices.SortFunc(candidateVersions, func(a, b *semver.Version) int {
		return a.Compare(b)
	})
	if len(candidateVersions) > 0 {
		if max {
			return candidateVersions[len(candidateVersions)-1].Original()
		}
		return candidateVersions[0].Original()
	}
	return ""
}

func getPackageEnginesNodeString(pkg string, versionRange string) (string, error) {
	var client http.Client
	packageVersions, err := getPackageVersions(pkg)
	if err != nil {
		return "", err
	}
	pkgVersion := getSuitableVersionString(Max, versionRange, packageVersions)

	enginesResponse, err := client.Get(fmt.Sprintf("%s/%s/%s", npmRegistryBaseUrl, pkg, pkgVersion))
	if err != nil {
		return "", err
	}
	defer enginesResponse.Body.Close()

	if enginesResponse.StatusCode == http.StatusOK {
		pkgInfo, err := io.ReadAll(enginesResponse.Body)
		if err != nil {
			return "", err
		}

		var pkgInfoJson PackageInfo
		err = json.Unmarshal(pkgInfo, &pkgInfoJson)
		if err != nil {
			return "", err
		}

		if len(pkgInfoJson.Engines) > 0 && pkgInfoJson.Engines["node"] != nil {
			pkgEngines := pkgInfoJson.Engines["node"].(string)
			fmt.Fprintln(out, bold(blue("[+]   ")), yellow("found"), "Node@"+pkgEngines, yellow("for"), pkg+"@"+pkgVersion)
			return pkgEngines, nil
		}
	}
	return "", err
}

func parsePackageJsonDependencies() (map[string]string, error) {
	dependencies := make(map[string]string)
	cwd, err := os.Getwd()
	if err != nil {
		return dependencies, err
	}
	packageJsonPath := cwd + "/package.json"
	f, err := os.Open(packageJsonPath)
	if err != nil {
		return dependencies, fmt.Errorf(bold(red("[!]")), "parsePackageJsonDependencies: %s %s", red("file not found:"), blue(packageJsonPath))
	}
	defer f.Close()

	fmt.Fprintln(out, bold(blue("[+] ")), yellow("parsing"), "package.json", yellow("dependencies"))
	packageJsonBinary, err := io.ReadAll(f)
	if err != nil {
		return dependencies, err
	}
	var packageJsonContent PackageJson
	json.Unmarshal(packageJsonBinary, &packageJsonContent)

	for dep, ver := range packageJsonContent.Dependencies {
		dependencies[dep] = ver
	}
	for dep, ver := range packageJsonContent.DevDependencies {
		dependencies[dep] = ver
	}

	return dependencies, nil
}

func convertRangeStringsToConstraints(engines Set[string]) []*semver.Constraints {
	constraints := make([]*semver.Constraints, 0, len(engines))
	for engine := range engines {
		engineSemverConstraint, err := semver.NewConstraint(engine)
		if err != nil {
			fmt.Fprintln(out, bold(red("[!]")), engine, "is not a valid semver")
			continue
		}
		constraints = append(constraints, engineSemverConstraint)
	}
	return constraints
}

func getSatisfyingVersions(semvers []*semver.Version, constraints []*semver.Constraints) []*semver.Version {
	satisfyingSemvers := make([]*semver.Version, 0, 10)
	for _, nodeSemver := range semvers {
		isCandidate := true
		for _, constraint := range constraints {
			if !constraint.Check(nodeSemver) {
				isCandidate = false
				break
			}
		}
		if isCandidate {
			satisfyingSemvers = append(satisfyingSemvers, nodeSemver)
		}
	}
	slices.SortFunc(satisfyingSemvers, func(a, b *semver.Version) int {
		return a.Compare(b)
	})
	return satisfyingSemvers
}

func GetSatisfyingNodeVersion(max MinMax) string {
	dependencies, err := parsePackageJsonDependencies()
	check(err)

	engines := make(Set[string])
	queue := make(chan string, len(dependencies))
	var wg sync.WaitGroup
	for dep, ver := range dependencies {
		wg.Add(1)
		fmt.Fprintln(out, bold(blue("[+]")), yellow("fetching"), dep+"@"+ver, yellow("metadata"))

		go func(d, v string) {
			defer wg.Done()
			engine, err := getPackageEnginesNodeString(d, v)
			check(err)
			queue <- engine
		}(dep, ver)
	}
	wg.Wait()
	close(queue)
	for engine := range queue {
		engines.Add(engine)
	}
	engines.Remove("")

	engineConstraints := convertRangeStringsToConstraints(engines)
	allNodeSemvers, err := getAllNodeSemvers(flags.isNonLts)
	check(err)
	validNodeSemvers := getSatisfyingVersions(allNodeSemvers, engineConstraints)

	if max {
		return validNodeSemvers[len(validNodeSemvers)-1].Original()
	}
	return validNodeSemvers[0].Original()
}
