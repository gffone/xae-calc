package handler

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"proc/internal/proc"
	"proc/internal/storage"
	"strconv"
	"strings"
)

var (
	statusFlag bool

	start int
	end   int
)

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/index.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "index", nil)
}

func current(w http.ResponseWriter, r *http.Request) {
	if !statusFlag {
		tmpl, _ := template.ParseFiles("web/templates/current_start.html", "web/templates/header.html", "web/templates/footer.html")
		tmpl.ExecuteTemplate(w, "current_start", nil)

	} else {
		tmpl, _ := template.ParseFiles("web/templates/current.html", "web/templates/header.html", "web/templates/footer.html")
		tmpl.ExecuteTemplate(w, "current", nil)
	}
}

func currentGraph(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/current_graph.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "current_graph", nil)
}

func success(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/success.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "success", nil)
}

func lib(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/lib.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "lib", nil)
}

func history(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/history.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "history", nil)
}

func base(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/base.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "base", nil)
}

func favourites(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/favourites.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "favourites", nil)
}

func groupingMode(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/grouping_mode.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "grouping_mode", nil)
}

func groupingModeDone(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/grouping_mode_done.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "grouping_mode_done", nil)
}

func groupingModeDoneCurrent(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/templates/grouping_mode_done_current.html", "web/templates/header.html", "web/templates/footer.html")
	tmpl.ExecuteTemplate(w, "grouping_mode_done_current", nil)
}

func submit(log *slog.Logger, strg *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		start, err = strconv.Atoi(r.FormValue("FormStart"))
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		end, err = strconv.Atoi(r.FormValue("FormEnd"))
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		err = proc.StandardProc(float64(start), float64(end), strg)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		tmpl, _ := template.ParseFiles("web/templates/current.html", "web/templates/header.html", "web/templates/footer.html")
		tmpl.ExecuteTemplate(w, "current", nil)
		statusFlag = true
	}
}

func upload(log *slog.Logger, strg *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statusFlag = false

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := recreateTempDir(strg); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		err := r.ParseMultipartForm(32 << 40)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		file, _, err := r.FormFile("uploadfile")
		defer file.Close()
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		f, err := os.OpenFile(fmt.Sprintf("%s/files/file.xlsx", strg.StoragePath), os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		_, err = io.Copy(f, file)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
		}

		tmpl, _ := template.ParseFiles("web/templates/success.html", "web/templates/header.html", "web/templates/footer.html")
		tmpl.ExecuteTemplate(w, "success", nil)
	}
}
func readFromGroupList(log *slog.Logger, strg *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		strVals := make([]string, 0)
		i := 1
		for {
			formVal := r.FormValue(fmt.Sprintf("%d", i))
			if formVal == "" {
				break
			}
			strVals = append(strVals, formVal)
			i++
		}

		intVals := make([][]int, i)
		for idx, strVal := range strVals {
			strArr := strings.Split(strVal, ",")
			for _, strEl := range strArr {
				intVal, err := strconv.Atoi(strEl)
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					log.Error(err.Error())
					return
				}
				intVals[idx] = append(intVals[idx], intVal)
			}
		}

		if err := recreateTempDirGroup(strg); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		err := proc.GroupModeProc(intVals[:len(intVals)-1], float64(start), float64(end), strg)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		tmpl, err := template.ParseFiles("web/templates/grouping_mode_done.html", "web/templates/header.html", "web/templates/footer.html")
		err = tmpl.ExecuteTemplate(w, "grouping_mode_done", nil)
	}
}

func recreateTempDirGroup(strg *storage.Storage) error {

	chartsDir := fmt.Sprintf("%s/group_charts/", strg.StoragePath)

	os.RemoveAll(chartsDir)

	err := os.MkdirAll(chartsDir, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func recreateTempDir(strg *storage.Storage) error {

	os.RemoveAll(strg.StoragePath)

	err := os.MkdirAll(fmt.Sprintf("%s/files/", strg.StoragePath), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/charts/", strg.StoragePath), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/bar/", strg.StoragePath), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func HandlerFunctions(log *slog.Logger, strg *storage.Storage) {
	http.HandleFunc("/", current)
	http.HandleFunc("/index", index)
	http.HandleFunc("/lib", lib)
	http.HandleFunc("/history", history)
	http.HandleFunc("/base", base)
	http.HandleFunc("/favourites", favourites)
	http.HandleFunc("/upload", upload(log, strg))
	http.HandleFunc("/submit", submit(log, strg))
	http.HandleFunc("/current", current)
	http.HandleFunc("/grouping_mode", groupingMode)
	http.HandleFunc("/current_graph", currentGraph)
	http.HandleFunc("/grouping_mode_done", groupingModeDone)
	http.HandleFunc("/grouping_mode_done_current", groupingModeDoneCurrent)
	http.HandleFunc("/read_from_group_list", readFromGroupList(log, strg))
	http.HandleFunc("/success", success)
}
