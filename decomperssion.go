package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/mholt/archiver"
	"golang.org/x/sync/semaphore"
)

var wg sync.WaitGroup
var workDir string
var (
	OverwriteExisting = true
	ContinueOnError   = false
	maxWorkers        = runtime.GOMAXPROCS(0)
	sem               = semaphore.NewWeighted(int64(maxWorkers))
	ctx               = context.TODO()
)

func init() {
	flag.StringVar(&workDir, "w", "worker", "工作路径r")
	flag.BoolVar(&ContinueOnError, "e", false, "出错是否继续")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

func fileMd5(file_path string) (md5Code string, err error) {
	f, err := os.Open(file_path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	hasher := md5.New()

	_, err = io.Copy(hasher, r)
	if err != nil {
		return "", err
	}
	str := hex.EncodeToString(hasher.Sum(nil))
	return str, nil

}

func doPath(runPath string) {
	wg.Add(1)
	go func(runPath string) {
		defer wg.Done()
		log.Printf("Dir Start %s\n", runPath)
		fileChan := make(chan string, 1024)
		inputFileChan := chan<- string(fileChan)
		go getFiles(runPath, inputFileChan)
		for p := range fileChan {
			doFile(p)
		}
		log.Printf("Dir End %s\n", runPath)
	}(runPath)

}

func checkAndMoveFile(src string, target string) bool {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.MkdirAll(path.Dir(target), os.ModePerm)
		// log.Printf("move %s %s\n", src, target)
		err := os.Rename(src, target)
		if err != nil {
			log.Printf("file move error: %s\n", src)
		}
		return true
	} else {
		os.Remove(src)
		return false
	}

}

func checkRarExist(md5Code string, ext string) bool {
	fp := "rar/" + md5Code + ext
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		fp = "rar_error/" + md5Code + ext
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func doFile(filePath string) {
	sem.Acquire(ctx, 1)
	wg.Add(1)
	go func(filePath string) {
		defer wg.Done()
		defer sem.Release(1)
		if strings.HasSuffix(filePath, "downloading") || strings.HasSuffix(filePath, "downloading.cfg") || strings.HasSuffix(filePath, "download") {
			log.Printf("Ignore downloading file %s\n", filePath)
			return
		}
		md5Code, err := fileMd5(filePath)
		if err != nil {
			log.Println(err)
			return
		}
		ext := strings.ToLower(filepath.Ext(filePath))
		if ext == ".rar" {
			fp := "rar/" + md5Code + ext
			if checkRarExist(md5Code, ext) {
				// path/to/whatever does not exist
				r := archiver.NewRar()
				r.OverwriteExisting = true
				r.ContinueOnError = ContinueOnError
				log.Printf("RAR start %s \n", filePath)
				err := r.Unarchive(filePath, "unrar/"+md5Code)
				if err != nil {
					log.Printf("Extract %s error %s\n", filePath, err)
					fp = "rar_error/" + md5Code + ".rar"
					err1 := os.Rename(filePath, fp)
					if err1 != nil {
						log.Printf("file move error: %s\n", filePath)
					}
				}

				// os.MkdirAll(path.Dir(fp), os.ModePerm)
				// log.Printf("move %s %s\n", filePath, fp)
				err1 := os.Rename(filePath, fp)
				if err1 != nil {
					log.Printf("file move error: %s\n", filePath)
					return
				}
				doPath("unrar/" + md5Code)
				log.Printf("RAR end %s \n", filePath)
			} else {
				os.Remove(filePath)
			}
		} else if ext == ".zip" {
			fp := "rar/" + md5Code + ext
			if checkRarExist(md5Code, ext) {
				// path/to/whatever does not exist
				r := archiver.NewZip()
				r.OverwriteExisting = OverwriteExisting
				r.ContinueOnError = ContinueOnError
				log.Printf("RAR start %s \n", filePath)
				err := r.Unarchive(filePath, "unrar/"+md5Code)
				if err != nil {
					log.Printf("Extract %s error %s\n", filePath, err)
					fp = "rar_error/" + md5Code + ".zip"
					err1 := os.Rename(filePath, fp)
					if err1 != nil {
						log.Printf("file move error: %s\n", filePath)
					}
				}

				// os.MkdirAll(path.Dir(fp), os.ModePerm)
				// log.Printf("move %s %s\n", filePath, fp)
				err1 := os.Rename(filePath, fp)
				if err1 != nil {
					log.Printf("file move error: %s\n", filePath)
					return
				}
				doPath("unrar/" + md5Code)
				log.Printf("RAR end %s \n", filePath)
			} else {
				os.Remove(filePath)
			}
		} else if ext == ".tar" {
			fp := "rar/" + md5Code + ext
			if checkRarExist(md5Code, ext) {
				// path/to/whatever does not exist
				r := archiver.NewTar()
				r.OverwriteExisting = OverwriteExisting
				r.ContinueOnError = ContinueOnError
				log.Printf("RAR start %s \n", filePath)
				err := r.Unarchive(filePath, "unrar/"+md5Code)
				if err != nil {
					log.Printf("Extract %s error %s\n", filePath, err)
					fp = "rar_error/" + md5Code + ".tar"
					err1 := os.Rename(filePath, fp)
					if err1 != nil {
						log.Printf("file move error: %s\n", filePath)
					}
				}

				// os.MkdirAll(filepath.Dir(fp), os.ModePerm)
				// log.Printf("move %s %s\n", filePath, fp)
				err1 := os.Rename(filePath, fp)
				if err1 != nil {
					log.Printf("file move error: %s\n", filePath)
					return
				}
				doPath("unrar/" + md5Code)
				log.Printf("RAR end %s \n", filePath)
			} else {
				os.Remove(filePath)
			}
		} else if strings.HasSuffix(filePath, ".tar.gz") {
			fp := "rar/" + md5Code + ".tar.gz"
			if checkRarExist(md5Code, ".tar.gz") {
				// path/to/whatever does not exist
				r := archiver.NewTarGz()
				r.OverwriteExisting = OverwriteExisting
				r.ContinueOnError = ContinueOnError
				log.Printf("RAR start %s \n", filePath)
				err := r.Unarchive(filePath, "unrar/"+md5Code)
				if err != nil {
					log.Printf("Extract %s error %s\n", filePath, err)
					fp = "rar_error/" + md5Code + ".tar.gz"
					err1 := os.Rename(filePath, fp)
					if err1 != nil {
						log.Printf("file move error: %s\n", filePath)
					}
				}

				// os.MkdirAll(filepath.Dir(fp), os.ModePerm)
				// log.Printf("move %s %s\n", filePath, fp)
				err1 := os.Rename(filePath, fp)
				if err1 != nil {
					log.Printf("file move error: %s\n", filePath)
					return
				}
				doPath("unrar/" + md5Code)
				log.Printf("RAR end %s \n", filePath)
			} else {
				os.Remove(filePath)
			}
		} else if ext == ".7z" {
			fp := "rar/" + md5Code + ext
			if checkRarExist(md5Code, ext) {
				// path/to/whatever does not exist
				log.Printf("RAR start %s \n", filePath)
				cmd := exec.Command("7z", "x", "-p", "-y", "-ounrar/"+md5Code, filePath)
				out, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("cmd.Run() failed with %s\n", err)
					log.Printf("combined out:\n%s\n", string(out))
					fp = "rar_error/" + md5Code + ".7z"
					err1 := os.Rename(filePath, fp)
					if err1 != nil {
						log.Printf("file move error: %s\n", filePath)
					}
					return
				}
				log.Printf("combined out:\n%s\n", string(out))

				// log.Printf("move %s %s\n", filePath, fp)
				err1 := os.Rename(filePath, fp)
				if err1 != nil {
					log.Printf("file move error: %s\n", filePath)
					return
				}
				doPath("unrar/" + md5Code)
				log.Printf("RAR end %s \n", filePath)
			} else {
				os.Remove(filePath)
			}
		} else if ext == ".doc" || ext == ".docx" || ext == ".htm" || ext == ".html" || ext == ".txt" || ext == ".mht" || ext == ".eml" || ext == ".pdf" {
			fp := "doc/" + md5Code[:2] + "/" + md5Code + filepath.Ext(filePath)
			checkAndMoveFile(filePath, fp)
		} else if ext == ".xls" || ext == ".xlsx" {
			fp := "list/" + md5Code + filepath.Ext(filePath)
			checkAndMoveFile(filePath, fp)
		} else {
			fp := "other/" + md5Code + filepath.Ext(filePath)
			checkAndMoveFile(filePath, fp)
		}
	}(filePath)

}

func cleanEmptyDirs(rootpath string) {
	log.Println("clean", rootpath)
	dirNames := make([]string, 0)
	err := filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			dirNames = append(dirNames, path)
		}
		return nil
	})
	if err != nil {
		log.Printf("walk error [%v]\n", err)
	}
	sort.Sort(sort.StringSlice(dirNames))
	for i := len(dirNames) - 1; i >= 0; i-- {
		s, err1 := ioutil.ReadDir(dirNames[i])
		// log.Println(dirNames[i])
		if err1 == nil && len(s) == 0 {

			os.Remove(dirNames[i])
		}
	}
}

func getFiles(rootpath string, fileChan chan<- string) {
	defer close(fileChan)
	err := filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		fileChan <- path
		return nil
	})
	if err != nil {
		log.Printf("walk error [%v]\n", err)
	}
}

func help() {

}

func main() {
	flag.Parse()
	// log.SetOutput(f)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Println("为获取到当前路径")
		os.Exit(1)
	}
	if _, err := os.Stat(dir + "/" + workDir); os.IsNotExist(err) {
		log.Println("工作文件夹不存在" + dir + "/" + workDir)
		os.Exit(1)
	}

	os.MkdirAll("unrar", os.ModePerm)
	os.MkdirAll("rar", os.ModePerm)
	os.MkdirAll("rar_error", os.ModePerm)
	os.MkdirAll("list", os.ModePerm)
	os.MkdirAll("doc", os.ModePerm)
	os.MkdirAll("other", os.ModePerm)

	cleanEmptyDirs(dir + "/unrar")
	doPath(dir + "/unrar")
	doPath(dir + "/" + workDir)
	wg.Wait()
	cleanEmptyDirs(dir + "/unrar")
	cleanEmptyDirs(dir + "/" + workDir)
	log.Println("Done")
}
