package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

var getAllNodeVersionsUrl string = "https://nodejs.org/dist/index.json"
var npmRegistryBaseUrl string = "https://registry.npmjs.com"

func GetAllNodeVersions() (map[string]string, error) {
	var nodeVersions []NodeVersion
	var nodeNpmVersionMap = make(map[string]string)
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
		nodeNpmVersionMap[node.Version] = node.Npm
	}
	return nodeNpmVersionMap, err
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

func getSuitableVersion(versionRange string, depVersions []string) string {
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

	max, err := semver.NewVersion("0.0.0")
	if err != nil {
		return candidateVersions[len(candidateVersions)-1].Original()
	}
	for _, candidateVersion := range candidateVersions {
		if candidateVersion.GreaterThan(max) {
			max = candidateVersion
		}
	}

	return max.Original()
}

func GetPackageEnginesNode(pkg string, versionRange string) (string, error) {
	var client http.Client
	packageVersions, err := getPackageVersions(pkg)
	if err != nil {
		return "", err
	}
	pkgVersion := getSuitableVersion(versionRange, packageVersions)

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
		if len(pkgInfoJson.Engines) > 0 {
			pkgEngines := pkgInfoJson.Engines["node"].(string)
			fmt.Fprintln(out, bold(blue("[+]   ")), yellow("found"), "Node@"+pkgEngines, yellow("for"), pkg+"@"+pkgVersion)
			return pkgEngines, nil
		}
	}
	return "", err
}

func GetPackageJsonDependencies() (map[string]string, error) {
	dependencies := make(map[string]string)
	cwd, err := os.Getwd()
	if err != nil {
		return dependencies, err
	}
	packageJsonPath := cwd + "/package.json"
	f, err := os.Open(packageJsonPath)
	if err != nil {
		return dependencies, fmt.Errorf(bold(red("[!]")), "GetPackageJsonDependencies: %s %s", red("file not found:"), blue(packageJsonPath))
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

func GetMinNodeVersion() string {
	dependencies, err := GetPackageJsonDependencies()
	check(err)

	engines := make(Set[string])
	var wg sync.WaitGroup
	for dep, ver := range dependencies {
		wg.Add(1)
		fmt.Fprintln(out, bold(blue("[+]")), yellow("fetching"), dep+"@"+ver, yellow("metadata"))

		go func(d, v string) {
			defer wg.Done()
			engine, err := GetPackageEnginesNode(d, v)
			check(err)
			engines.Add(engine)
		}(dep, ver)
	}
	wg.Wait()
	engines.Remove("")

	engineSemverConstraints := make([]*semver.Constraints, 0, len(engines))
	for engine := range engines {
		engineSemverConstraint, err := semver.NewConstraint(engine)
		if err != nil {
			fmt.Fprintln(out, bold(red("[!]")), engine, "is not a valid semver")
			continue
		}
		engineSemverConstraints = append(engineSemverConstraints, engineSemverConstraint)
	}

	allNodeVersions, err := GetAllNodeVersions()
	check(err)
	candidateNodeSemvers := make([]*semver.Version, 0, 10)
	for nodeVersion := range allNodeVersions {
		nodeSemver, err := semver.NewVersion(nodeVersion)
		if err != nil {
			fmt.Fprintln(out, bold(red("[!]")), nodeVersion, "is not a valid semver")
			continue
		}
		isCandidate := true
		for _, constraint := range engineSemverConstraints {
			if !constraint.Check(nodeSemver) {
				isCandidate = false
				break
			}
		}
		if isCandidate {
			candidateNodeSemvers = append(candidateNodeSemvers, nodeSemver)
		}
	}

	if len(candidateNodeSemvers) > 0 {
		minNodeSemver := candidateNodeSemvers[0]
		for _, nodeSemver := range candidateNodeSemvers {
			if nodeSemver.LessThan(minNodeSemver) {
				minNodeSemver = nodeSemver
			}
		}
		return minNodeSemver.Original()
	}

	return "0.0.0"
}
