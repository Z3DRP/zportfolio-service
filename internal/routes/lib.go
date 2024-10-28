package routes

import (
	"fmt"
	"net/http"
	"strings"
)

// TODO refactor methods to accept a log level

func handleConfigReadErr(e error, w http.ResponseWriter) {
	emsg := fmt.Sprintf("%v", e)
	log(emsg, "debug")
	http.Error(w, emsg, http.StatusInternalServerError)
}

func handleRedisConErr(e error, w http.ResponseWriter) {
	emsg := fmt.Sprintf("%v", e)
	log(emsg, "debug")
	http.Error(w, emsg, http.StatusInternalServerError)
}

func handleJsonDecodeErr(operation string, e error, w http.ResponseWriter) {
	derr := NewDecodeRequestBodyErr(operation, e)
	emsg := fmt.Sprintf("%v", derr)
	log(emsg, "debug")
	http.Error(w, emsg, http.StatusInternalServerError)
}

func handleJsonEncodeErr(stype string, e error, w http.ResponseWriter) {
	emsg := fmt.Sprintf("could not encode %v response into json:: %v", stype, e)
	log(emsg, "debug")
	http.Error(w, emsg, http.StatusInternalServerError)
}

func handleTaskTimeParseErr(timeType string, e error, w http.ResponseWriter) {
	emsg := fmt.Sprintf("invalid type for task %v date:: %v", timeType, e)
	log(emsg, "debug")
	http.Error(w, emsg, http.StatusBadRequest)
}

func handleCacheReadErr(cacheType string, e error, w http.ResponseWriter) {
	emsg := fmt.Sprintf("error occurred while reading %v cache:: %v", cacheType, e)
	log(emsg, "debug")
	http.Error(w, emsg, http.StatusInternalServerError)
}

func handleIdGeneratorErr(idType string, e error, w http.ResponseWriter) {
	emsg := fmt.Sprintf("could not generate %v id:: %v", idType, e)
	log(emsg, "debug")
	http.Error(w, emsg, http.StatusInternalServerError)
}

func handleCacheSetErr(cType string, e error, w http.ResponseWriter) {
	emsg := fmt.Sprintf("error occurred while creating %v cache:: %v", cType, e)
	log(emsg, "debug")
	http.Error(w, emsg, http.StatusInternalServerError)
}

func handleTaskActionErr(action string, e error, uid string, w http.ResponseWriter) {
	emsg := fmt.Sprintf("error occurred while usr: %v tried %v task:: %v", uid, action, e)
	log(emsg, "debug")
	http.Error(w, fmt.Sprintf("error occurred while %v task:: %v", action, e), http.StatusInternalServerError)
}

func handleTypeCaseErr(typ string, t interface{}, e error, w http.ResponseWriter) {
	emsg := fmt.Sprintf("could not cast Type[%T] as %v:: %v", t, typ, e)
	log(emsg, "debug")
	http.Error(w, emsg, http.StatusInternalServerError)
}

func handleRequestTimeout(r *http.Request, w http.ResponseWriter) {
	re := NewRequestTimeoutErr(r)
	log(fmt.Sprintf("%v", re), "debug")
	http.Error(w, "request time out", http.StatusInternalServerError)
}

func log(m, lvl string) {
	switch strings.ToLower(lvl) {
	case "debug":
		logger.MustDebug(m)
	default:
		logger.MustTrace(m)
	}
}
