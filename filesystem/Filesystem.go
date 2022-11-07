package filesystem

import (
	"errors"
	"fmt"
	"github.com/shammishailaj/chief/system"
	log "github.com/sirupsen/logrus"
	"gopkg.in/djherbis/times.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DEL_FILE_LIST_E_SLEEP = 60
)

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func DeleteFileByAge(path string, minAgeForDeletion int64) (bool, error) {
	log.Infof("Deleting file: %s", path)
	/*fileStat*/ _, fileStatErr := os.Stat(path)
	if os.IsNotExist(fileStatErr) {
		log.Errorf("Could not find file %s. %s", path, fileStatErr.Error())
		return false, fileStatErr
	} else {
		t, err := times.Stat(path)
		if err != nil {
			log.Errorf("Error stating file %s using times.Stat(). %s", path, err.Error())
			return false, err
		}
		//tFileCreationTime := fileStat.Sys().(*syscall.Stat_t).Ctimespec
		//fileCreationTime := time.Unix(tFileCreationTime.Sec, tFileCreationTime.Nsec)
		var fileCreationTime time.Time
		if t.HasBirthTime() {
			fileCreationTime = t.BirthTime()
		}
		if t.HasChangeTime() {
			fileCreationTime = t.ChangeTime()
		}
		fileModTime := t.ModTime()
		fileAccessTime := t.AccessTime()

		tCurrentTime := time.Now()
		tCurrentTimeUnix := tCurrentTime.Unix()
		tFileAgeForDeletion := minAgeForDeletion // 10 secs X 60 = 600 secs  OR 10 mins

		if (tCurrentTimeUnix - fileCreationTime.Unix()) >= tFileAgeForDeletion {
			delFileErr := os.Remove(path)
			if delFileErr != nil {
				log.Errorf("FAILED to remove file %s. %s", path, delFileErr.Error())
				return false, delFileErr
			} else {
				log.Infof("Successfully Removed file %s", path)
				return true, nil
			}
		} else {
			log.Errorf("Specified File: %s is newer than specified deletion age of - %d second(s). WONT DELETE!", path, minAgeForDeletion)
			log.Errorf("File: %s, Current Time (Unix): %d, Creation Time (Unix): %d, Age for Deletion: %d", path, tCurrentTimeUnix, fileCreationTime.Unix(), tFileAgeForDeletion)
			log.Errorf("Creation Time: %s, Modification Time: %s, Access Time: %s", fileCreationTime.String(), fileModTime.String(), fileAccessTime.String())
			return false, errors.New(fmt.Sprintf("Specified File: %s is newer than specified deletion age of - %d second(s). WONT DELETE!", path, minAgeForDeletion))
		}
	}
}

func GetFileList(directoryPath string) map[int]string {
	filesList := make(map[int]string)
	filesListIterator := 0

	log.Infof("directoryPath = %s", directoryPath)
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		log.Errorf("Error reading directories: %s", err.Error())
		return filesList
	}

	for fileKey, file := range files {
		fileStat, fileStatErr := os.Stat(file.Name())
		if fileStatErr != nil {
			log.Errorf("Error stating file #%d: %s. %s", fileKey, file.Name(), fileStatErr.Error())
			continue
		}

		if fileStat.IsDir() {
			tFilesList := GetFileList(file.Name())
			tFilesListLen := len(tFilesList)
			if tFilesListLen > 0 {
				for _, value := range tFilesList {
					filesList[filesListIterator] = value
					filesListIterator++
				}
			}
			continue
		}

		filesList[filesListIterator] = file.Name()
		filesListIterator++
	}
	return filesList
}

func GetFileListGlob(pattern string) map[int]string {
	filesList := make(map[int]string)
	filesListIterator := 0

	log.Infof("Globbing pattern: %s", pattern)
	files, filesErr := filepath.Glob(pattern)
	if filesErr != nil {
		log.Errorf("Error reading directories: %s", filesErr.Error())
		return filesList
	}

	for fileKey, file := range files {
		fileStat, fileStatErr := os.Stat(file)
		if fileStatErr != nil {
			log.Errorf("Error stating file #%d: %s. %s", fileKey, file, fileStatErr.Error())
			continue
		}

		if fileStat.IsDir() {
			log.Printf("Pattern = %s", pattern)
			patternBase := filepath.Base(pattern)
			log.Printf("Pattern Base = %s", patternBase)
			patternPath := strings.TrimSuffix(pattern, patternBase)
			log.Printf("pattern path = %s", patternPath)
			newDirWithPattern := fmt.Sprintf("%s/%s", file, patternBase)
			log.Printf("New Dir Pattern = %s", newDirWithPattern)

			tFilesList := GetFileListGlob(newDirWithPattern)
			tFilesListLen := len(tFilesList)
			if tFilesListLen > 0 {
				for _, value := range tFilesList {
					filesList[filesListIterator] = value
					filesListIterator++
				}
			}
			continue
		}

		filesList[filesListIterator] = file
		filesListIterator++
	}
	return filesList
}

func DeleteFileList(fileExtToClean, directoryPath string) int {
	deletionCount := 0
	if fileExtToClean == "" {
		fileExtToClean = "pdf"
	}

	fileExtToClean = fmt.Sprintf("*%s", fileExtToClean)

	if directoryPath == "" {
		log.Errorf("Must specify Directory Path where files need to deleted")
	} else {
		if FileExists(directoryPath) {
			filesList := GetFileList(fmt.Sprintf("%s/%s", directoryPath, fileExtToClean))
			filesListLen := len(filesList)
			if filesListLen > 0 {
				for _, filePath := range filesList {
					fileDel, fileDelErr := DeleteFileByAge(filePath, 600)
					if fileDelErr == nil {
						if fileDel {
							deletionCount++
						}
					} else {
						log.Errorf("Error deleting file %s. %s", filePath, fileDelErr.Error())
					}
				}
			} else {
				log.Infof("No files found in directory %s", directoryPath)
			}
		} else {
			log.Infof("Path %s does not exist", directoryPath)
		}
	}

	log.Infof("Deleted %d %s files at path %s", deletionCount, fileExtToClean, directoryPath)
	return deletionCount
}

func DeleteFileListE(prefix, fileExtToClean, directoryPath string, maxAge int64, forceNoExt bool) int {
	deletionCount := 0
	cpuCores := system.CPUCores()

	log.Infof("LoadAvgCheck():: Found %d CPU Cores", cpuCores)
	if !forceNoExt && fileExtToClean == "" {
		log.Infof("Extensionless option is not being force with an empty/nil file extension. Defaulting to .PDF")
		fileExtToClean = "pdf"
	}
	globPatternFormat := "%s*.%s"
	if forceNoExt {
		log.Infof("--force-no-ext is true, removing . (dot) from globPatternFormat")
		globPatternFormat = "%s*%s"
	}

	fileExtToClean = fmt.Sprintf(globPatternFormat, prefix, fileExtToClean)

	if directoryPath == "" {
		log.Errorf("Must specify Directory Path where files need to deleted")
	} else {
		if FileExists(directoryPath) {
			filesList := GetFileListGlob(fmt.Sprintf("%s/%s", directoryPath, fileExtToClean))
			filesListLen := len(filesList)
			log.Infof("Found %d files", filesListLen)
			if filesListLen > 0 {
				for _, filePath := range filesList {
					if system.LoadAvgCheckCPUCores(cpuCores) == system.LAVG_TREND_NORMAL {
						fileDel, fileDelErr := DeleteFileByAge(filePath, maxAge)
						if fileDelErr == nil {
							if fileDel {
								deletionCount++
							}
						} else {
							log.Errorf("Error deleting file %s. %s", filePath, fileDelErr.Error())
						}
					} else {
						loadAvg, loadAvgErr := system.LoadAvg()
						if loadAvgErr != nil {
							log.Errorf("Unable to read System Load Average. %s", loadAvgErr.Error())
						} else {
							log.Infof("Load Average (1), (5), (15) = (%f), (%f), (%f)", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)
						}
						log.Infof("Deleted %d file(s) till now", deletionCount)
						log.Infof("Sleeping for %d seconds...", DEL_FILE_LIST_E_SLEEP)
						time.Sleep(DEL_FILE_LIST_E_SLEEP * time.Second)
						log.Infof("Woke-up!!!")
					}
				}
			} else {
				log.Infof("No files found in directory %s", directoryPath)
			}
		} else {
			log.Infof("Path %s does not exist", directoryPath)
		}
	}

	log.Infof("Deleted %d %s files at path %s", deletionCount, fileExtToClean, directoryPath)
	return deletionCount
}
