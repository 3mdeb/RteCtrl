package restServer

import (
	"3mdeb/RteCtrl/pkg/flashromControl"
	iface "3mdeb/RteCtrl/pkg/gpiocontrol/iface"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type RestServer struct {
	RestPrefix  string
	RomFilename string
	Gpio        iface.IGpio
}

var flash *flashromControl.Flashrom
var tempRomFile string

type errorStatus struct {
	Error string `json:"error"`
}

/* error status strings */
const (
	errCantReadGpioStates      = "can't read GPIO states"
	errGpioIDNotFound          = "id not found"
	errBadParams               = "bad parameters"
	errCantUploadFile          = "can't upload file"
	errNoFileUploaded          = "no file uploaded"
	errChecksumMismatch        = "checksum mismatch"
	errNoFlashingPerformed     = "no flashing operation performed"
	errFlashVerificationFailed = "flash verification failed"
	errInternalError           = "internal error"
)

type gpioState struct {
	ID          int    `json:"id"`
	State       uint   `json:"state"`
	Description string `json:"description"`
	Direction   string `json:"direction"`
}

type gpioStateRequest struct {
	State     uint   `json:"state"`
	Direction string `json:"direction"`
	Time      uint   `json:"time"`
}

type status struct {
	Status  string `json:"status"`
	Percent uint   `json:"percent,omitempty"`
}

const (
	statStarting           = "starting"
	statPreparing          = "preparing"
	statReadingOldContents = "reading old contents"
	statErasingAndWriting  = "erasing flash and writing"
	statVerifying          = "verifying"
	statDone               = "done"
	statOk                 = "ok"
)

type flasherConfig struct {
	Speed int `json:"speed"`
}

type fileDetails struct {
	Checksum string `json:"file_md5"`
	Size     int64  `json:"file_size,omitempty"`
}

var currentFileDetails = fileDetails{
	Checksum: "",
	Size:     0,
}

func (server *RestServer) listAllGpios(w http.ResponseWriter, r *http.Request) {
	gpios := make([]gpioState, server.Gpio.GetNumberOfGpios())
	var err error
	out := json.NewEncoder(w)

	for i := range gpios {
		gpios[i].ID = i
		gpios[i].Description = server.Gpio.GetDescription(i)
		gpios[i].Direction, err = server.Gpio.GetDirection(i)
		if err != nil {
			break
		}
		gpios[i].State, err = server.Gpio.GetState(i)
		if err != nil {
			break
		}
	}

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errCantReadGpioStates})
		return
	}

	out.Encode(gpios)
}

func (server *RestServer) getGpioState(w http.ResponseWriter, r *http.Request) {
	out := json.NewEncoder(w)

	params := mux.Vars(r)

	log.Println("getGpioState:", params["id"])

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errGpioIDNotFound})
		return
	}

	var g gpioState

	g.ID = id
	g.Description = server.Gpio.GetDescription(id)
	g.Direction, err = server.Gpio.GetDirection(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errCantReadGpioStates})
		return
	}
	g.State, err = server.Gpio.GetState(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errCantReadGpioStates})
		return
	}

	out.Encode(g)
}

func (server *RestServer) setGpioState(w http.ResponseWriter, r *http.Request) {
	out := json.NewEncoder(w)
	params := mux.Vars(r)

	log.Println("setGpioState:", params["id"])

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errGpioIDNotFound})
		return
	}

	var g gpioStateRequest

	in := json.NewDecoder(r.Body)
	err = in.Decode(&g)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		out.Encode(errorStatus{Error: errBadParams})
		return
	}

	if g.Direction != "" {
		err = server.Gpio.SetDirection(id, g.Direction)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			out.Encode(errorStatus{Error: errCantReadGpioStates})
			return
		}
	}

	err = server.Gpio.SetState(id, g.State)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errCantReadGpioStates})
		return
	}

	if g.Time != 0 {
		timer := time.NewTimer(time.Second * time.Duration(g.Time))
		log.Println("setGpioState: timer started")
		go func() {
			<-timer.C
			log.Println("setGpioState: timer fired")
			if g.State == 0 {
				server.Gpio.SetState(id, 1)
			} else {
				server.Gpio.SetState(id, 0)
			}
		}()
	}
	server.getGpioState(w, r)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	out := json.NewEncoder(w)
	rf, _, err := r.FormFile("file")
	defer rf.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errCantUploadFile})
		return
	}
	h := md5.New()

	tee := io.TeeReader(rf, h)

	zr, err := gzip.NewReader(tee)
	if err != nil {
		log.Println("not gzip", err)
		rf.Seek(0, 0)
		h.Reset()
	}

	f, err := os.OpenFile(tempRomFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errCantUploadFile})
		return
	}
	defer f.Close()

	var written int64
	if zr != nil {
		written, err = io.Copy(f, zr)
	} else {
		written, err = io.Copy(f, tee)
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errCantUploadFile})
		return
	}

	currentFileDetails.Size = written
	currentFileDetails.Checksum = fmt.Sprintf("%x", h.Sum(nil))

	log.Print("file uploaded, size: ",
		currentFileDetails.Size,
		" ,md5: ",
		currentFileDetails.Checksum)

	out.Encode(currentFileDetails)
}

func getFileDetails(w http.ResponseWriter, r *http.Request) {
	out := json.NewEncoder(w)

	if currentFileDetails.Checksum == "" {
		log.Println("no file uploaded")
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errNoFileUploaded})
		return
	}

	out.Encode(currentFileDetails)
}

func removeFile(w http.ResponseWriter, r *http.Request) {
	out := json.NewEncoder(w)

	if currentFileDetails.Checksum == "" {
		log.Println("no file uploaded")
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errNoFileUploaded})
		return
	}

	os.Remove(tempRomFile)
	currentFileDetails.Checksum = ""
	currentFileDetails.Size = 0

	out.Encode(status{Status: statOk})
}

func getFlasherConfig(w http.ResponseWriter, r *http.Request) {
	out := json.NewEncoder(w)
	cfg := flash.GetConfig()

	out.Encode(flasherConfig{Speed: cfg.Speed})
}

func setFlasherConfig(w http.ResponseWriter, r *http.Request) {
	out := json.NewEncoder(w)
	in := json.NewDecoder(r.Body)
	var cfg flasherConfig

	err := in.Decode(&cfg)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		out.Encode(errorStatus{Error: errBadParams})
		return
	}

	flash.SetConfig(flashromControl.Config{Speed: cfg.Speed})

	getFlasherConfig(w, r)
}

func startFlashing(w http.ResponseWriter, r *http.Request) {
	out := json.NewEncoder(w)
	in := json.NewDecoder(r.Body)

	if currentFileDetails.Checksum == "" {
		log.Println("no file uploaded")
		w.WriteHeader(http.StatusNotFound)
		out.Encode(errorStatus{Error: errNoFileUploaded})
		return
	}

	var fDetails fileDetails
	err := in.Decode(&fDetails)
	if err != nil || fDetails.Checksum == "" {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		out.Encode(errorStatus{Error: errBadParams})
		return
	}

	if fDetails.Checksum != currentFileDetails.Checksum {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		out.Encode(errorStatus{Error: errChecksumMismatch})
		return
	}

	err = flash.Start(tempRomFile)
	if err == flashromControl.ErrProcessAlreadyStarted {
		getFlashingState(w, r)
		return
	} else if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		out.Encode(errorStatus{Error: errInternalError})
		return
	}

	out.Encode(status{Status: statStarting, Percent: 1})
}

func getFlashingState(w http.ResponseWriter, r *http.Request) {
	out := json.NewEncoder(w)

	state := flash.GetState()

	var s status

	switch state {
	case flashromControl.StateIdle:
		out.Encode(errorStatus{Error: errNoFlashingPerformed})
		return
	case flashromControl.StatePreparing:
		s.Status = statPreparing
		s.Percent = 20
	case flashromControl.StateReading:
		s.Status = statReadingOldContents
		s.Percent = 40
	case flashromControl.StateErasing:
		s.Status = statErasingAndWriting
		s.Percent = 60
	case flashromControl.StateVerifying:
		s.Status = statVerifying
		s.Percent = 80
	case flashromControl.StateDone:
		s.Status = statDone
		s.Percent = 100
	case flashromControl.StateError:
		out.Encode(errorStatus{Error: errFlashVerificationFailed})
		return
	default:
		out.Encode(errorStatus{Error: errBadParams})
		return
	}

	out.Encode(s)
}

func logFunc(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		var fName string = runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		fName = strings.Join(strings.Split(fName, ".")[1:], "")

		log.Printf("request [%s]: %s", r.Method, fName)
		f(w, r)
	}
}

func (server *RestServer) Start(address, webDir string, f *flashromControl.Flashrom) {

	flash = f

	tempRomFile = fmt.Sprintf("%s/%s", os.TempDir(), server.RomFilename)

	fs := http.FileServer(http.Dir(webDir))

	files, err := ioutil.ReadDir(webDir)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.Handle("/", fs).Methods("GET")
	for _, file := range files {
		router.Handle("/"+file.Name(), fs).Methods("GET")
	}

	router.HandleFunc(server.RestPrefix+"/gpio", logFunc(server.listAllGpios)).Methods("GET")
	router.HandleFunc(server.RestPrefix+"/gpio/{id}", logFunc(server.getGpioState)).Methods("GET")
	router.HandleFunc(server.RestPrefix+"/gpio/{id}", logFunc(server.setGpioState)).Methods("PATCH")

	router.HandleFunc(server.RestPrefix+"/flash/file", logFunc(uploadFile)).Methods("POST")
	router.HandleFunc(server.RestPrefix+"/flash/file", logFunc(getFileDetails)).Methods("GET")
	router.HandleFunc(server.RestPrefix+"/flash/file", logFunc(removeFile)).Methods("DELETE")

	router.HandleFunc(server.RestPrefix+"/flash/config", logFunc(getFlasherConfig)).Methods("GET")
	router.HandleFunc(server.RestPrefix+"/flash/config", logFunc(setFlasherConfig)).Methods("PATCH")
	router.HandleFunc(server.RestPrefix+"/flash", logFunc(startFlashing)).Methods("PUT")
	router.HandleFunc(server.RestPrefix+"/flash", logFunc(getFlashingState)).Methods("GET")

	http.ListenAndServe(address, router)
}
