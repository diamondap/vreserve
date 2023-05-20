package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

// VolumeService keeps track of the space available to workers
// processing APTrust bags.
type VolumeService struct {
	host    string
	port    int
	volumes map[string]*Volume
	logger  *logging.Logger
}

// NewVolumeService creates a new VolumeService object to track the
// amount of available space and claimed space on locally mounted
// volumes.
func NewVolumeService(host string, port int, logger *logging.Logger) *VolumeService {
	return &VolumeService{
		host:    host,
		port:    port,
		volumes: make(map[string]*Volume),
		logger:  logger,
	}
}

// Serve starts an HTTP server, so the VolumeService can respond to
// requests from the VolumeClient(s). See the VolumeClient for available
// calls.
func (service *VolumeService) Serve() {
	http.HandleFunc("/reserve/", service.makeReserveHandler())
	http.HandleFunc("/release/", service.makeReleaseHandler())
	http.HandleFunc("/report/", service.makeReportHandler())
	http.HandleFunc("/ping/", service.makePingHandler())
	listenAddr := fmt.Sprintf("%s:%d", service.host, service.port)
	http.ListenAndServe(listenAddr, nil)
}

// Returns a Volume object with info about the volume at the specified
// mount point. The mount point should be the path to a disk or partition.
// For example, "/", "/mnt/data", etc.
func (service *VolumeService) getVolume(path string) *Volume {
	mountpoint, err := GetMountPointFromPath(path)
	if err != nil {
		mountpoint = "/"
		service.logger.Error("Cannot determine mountpoint of file '%s': %v",
			path, err)
	}
	if _, keyExists := service.volumes[mountpoint]; !keyExists {
		service.volumes[mountpoint] = NewVolume(mountpoint)
	}
	return service.volumes[mountpoint]
}

func (service *VolumeService) makeReserveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &VolumeResponse{}
		status := http.StatusOK
		path := r.FormValue("path")
		bytes, err := strconv.ParseUint(r.FormValue("bytes"), 10, 64)
		if path == "" {
			response.Succeeded = false
			response.ErrorMessage = "Param 'path' is required."
			status = http.StatusBadRequest
		} else if err != nil || bytes < 1 {
			response.Succeeded = false
			response.ErrorMessage = "Param 'bytes' must be an integer greater than zero."
			status = http.StatusBadRequest
		} else {
			volume := service.getVolume(path)
			err = volume.Reserve(path, bytes)
			if err != nil {
				response.Succeeded = false
				response.ErrorMessage = fmt.Sprintf(
					"Could not reserve %d bytes for file '%s': %v",
					bytes, path, err)
				service.logger.Error("[%s] %s", r.RemoteAddr, response.ErrorMessage)
				status = http.StatusInternalServerError
			} else {
				response.Succeeded = true
				service.logger.Infof("[%s] Reserved %d bytes for %s", r.RemoteAddr, bytes, path)
			}
		}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(jsonResponse)
	}
}

func (service *VolumeService) makeReleaseHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &VolumeResponse{}
		path := r.FormValue("path")
		status := http.StatusOK
		if path == "" {
			response.Succeeded = false
			response.ErrorMessage = "Param 'path' is required."
			status = http.StatusBadRequest
		} else {
			volume := service.getVolume(path)
			volume.Release(path)
			response.Succeeded = true
			service.logger.Infof("[%s] Released %s", r.RemoteAddr, path)
		}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(jsonResponse)
	}
}

func (service *VolumeService) makeReportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &VolumeResponse{}
		path := r.FormValue("path")
		status := http.StatusOK
		if path == "" {
			response.Succeeded = false
			response.ErrorMessage = "Param 'path' is required."
			status = http.StatusBadRequest
		} else {
			volume := service.getVolume(path)
			response.Succeeded = true
			response.Data = volume.Reservations()
			service.logger.Infof("[%s] Reservations %s (%d)", r.RemoteAddr, path, len(response.Data))
		}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(jsonResponse)
	}
}

func (service *VolumeService) makePingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &VolumeResponse{}
		response.Succeeded = true
		status := http.StatusOK
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(jsonResponse)
	}
}

// On Linux and OSX, this uses df in a safe way (without passing
// through any user-supplied input) to find the mountpoint of a
// given file.
func GetMountPointFromPath(path string) (string, error) {
	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("windows is not supported :(")
	}
	out, err := exec.Command("df").Output()
	if err != nil {
		return "", err
	}
	matchingMountpoint := ""
	maxPrefixLen := 0
	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if i > 0 {
			words := strings.Split(line, " ")
			mountpoint := words[len(words)-1]
			if strings.HasPrefix(path, mountpoint) && len(mountpoint) > maxPrefixLen {
				matchingMountpoint = mountpoint
			}
		}
	}
	return matchingMountpoint, nil
}
