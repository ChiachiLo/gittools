package main
import "github.com/jacklo/gittools/git"
// import "github.com/jacklo/util"
import "io/ioutil"
import "fmt"
import "encoding/xml"
import "path"
import "log"
// import "path"
// import "path/filepath"
// import "os/exec"

// import "flag"
import "github.com/alexflint/go-arg"
import "strings"
import "os"

// import log "github.com/sirupsen/logrus"
// import "os"
// import "time"
// import "io/ioutil"

var args struct {
	OutputDir string           `arg:"required,-o"`
	ProjectRootDir *string     `arg:"-p" help:"Set project root dir."`
	ManifestFileList []string  `arg:"required,-m" help:"Set imput manifest file list."`
}

func parserArgs() {
	arg.MustParse(&args)
}

// <project name="A88/android/abl/tianocore/edk2" path="LINUX/android/bootable/bootloader/edk2" revision="9657d6522fed8c76f21934e8b4ad4c15072473b5" upstream="refs/heads/q35_la13_00042_pvt"/>
type Project struct {
	Name string `xml:"name,attr"`
	Path string `xml:"path,attr"`
	Revision string `xml:"revision,attr"`
}

type ProjectList struct {
	Projects []Project `xml:"project"`
}
func load_manifest(file_path string) *ProjectList {
	log.Printf("Load manifest file:%v", file_path)
	data, err := ioutil.ReadFile(file_path)
	if err != nil {
		log.Printf("Load manifest file fail..%v", file_path)
		return nil
	}
	project_list := ProjectList{}
	if err := xml.Unmarshal(data, &project_list); err != nil {
		log.Printf("Load manifest Unmarshal file fail..%v", file_path)
		log.Fatal(err)
		return nil
	}
	// log.Printf("manifest:%v", project_list)
	return &project_list
}

func conv_to_project_map(Projects *[]Project) map[string]string {
	log.Printf("conv_to_project_map")
	project_map := make(map[string]string)

	for _, value := range *Projects {
		project_map[value.Path] = value.Revision
	}
	// log.Printf("Project map:%v", project_map)
	return project_map
}

func create_project_diff_ary(proj1 map[string]string, proj2 map[string]string) []string {
	log.Printf("create_project_diff_ary")
	proj1_map:=clone_map(proj1)
  proj2_map:=clone_map(proj2)

	diff_proj := []string{}
	for key, val1 := range proj1_map {
		if val2, ok := proj2_map[key]; ok {
			if val1 == val2 {
				log.Printf("create_project_diff_ary 1")
				delete(proj2_map, key)
				continue
			}	else {
				diff_proj = append(diff_proj, key)
				delete(proj2_map, key)
			}
		}	else {
			diff_proj = append(diff_proj, key)
		}
	}

	for key, _ := range proj2_map {
		diff_proj = append(diff_proj, key)
	}
	// log.Printf("diff proj:\n%v", diff_proj)
	return diff_proj
}

func get_manifest_path(proj_root string, manifest_path string) string {
	if _, err := os.Stat(manifest_path); err == nil {
		return manifest_path
	}

	full_path := path.Join(proj_root, ".repo", "manifests", manifest_path)
	if _, err := os.Stat(full_path); err == nil {
		return full_path
	}

	log.Printf("Lookup %s not find", manifest_path)

	return ""
}

func clone_map(src_map map[string]string) map[string]string {
	dest_map := map[string]string{}
	for k, v := range src_map {
		dest_map[k] = v
	}
	return dest_map
}

func output_diff_folder(proj_root string, manifest_file_list []string, output_dir string) {

	// *ProjectList

	proj1_map := map[string]string{}
	for idx, _ := range manifest_file_list {
		if idx > (len(manifest_file_list) -2) {
			log.Printf("end to diff manifest.")
			return
		}

		// if idx == 0 {
		mf1 := load_manifest(get_manifest_path(proj_root, manifest_file_list[idx]))
		proj1_map = conv_to_project_map(&mf1.Projects)
		// }
		mf2 := load_manifest(get_manifest_path(proj_root, manifest_file_list[idx + 1]))
		proj2_map := conv_to_project_map(&mf2.Projects)

		diff_proj := create_project_diff_ary(proj1_map, proj2_map)
		log.Printf("create_project_diff_ary ok")

		for _, proj_path := range diff_proj {
			log.Printf("proj_path %v", proj_path)
			diff_path_1 := path.Join(output_dir, fmt.Sprintf("%v_%v",idx, path.Base( strings.ReplaceAll(manifest_file_list[idx],".","_"))),   proj_path)
			diff_path_2 := path.Join(output_dir, fmt.Sprintf("%v_%v",idx, path.Base( strings.ReplaceAll(manifest_file_list[idx+1],".","_"))), proj_path)
			git_folder  := path.Join(proj_root, proj_path)
			log.Printf("git_folder %v git1:%v, git2:%v", git_folder, proj1_map[proj_path], proj2_map[proj_path])
			// log.Printf("git_folder %v", git_folder)
			git_diff_file_list := git.GetDiffList(git_folder, proj1_map[proj_path], proj2_map[proj_path])
			log.Printf("git_diff_file_list %v", git_diff_file_list)
			// git_tree_map_1 := git.GetTreeMap(git_folder, proj1_map[proj_path])
			// git_tree_map_2 := git.GetTreeMap(git_folder, proj2_map[proj_path])
			git.TackOutFileFromList(git_folder, proj1_map[proj_path], git_diff_file_list, diff_path_1)
			git.TackOutFileFromList(git_folder, proj2_map[proj_path], git_diff_file_list, diff_path_2)
		}

		proj1_map = proj2_map
	}
}


func main() {
	parserArgs()
	// test()

	fmt.Printf("args:%v", args)
	if args.ProjectRootDir == nil {
		temp := "."
		args.ProjectRootDir = &temp
	}
	if len(args.ManifestFileList) < 2 {
		fmt.Printf("Need define more manifest file")
		os.Exit(1)
	}
	output_diff_folder(
		*args.ProjectRootDir,
		args.ManifestFileList,
		args.OutputDir)

}

// var args struct {
// 	OutputDir string           `arg:"-o"`
// 	ProjectRootDir *string      `arg:"-p" help:"Set project root dir."`
// 	ManifestFileList []string  `arg:"-m" help:"Set imput manifest file list."`
// }