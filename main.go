package main

import (

	"fmt"
	"html/template"
	"io"
	"proc/proc"
	"strings"
	"net/http"
	"os"
	"strconv"
)

var checking_if_new_processing_has_started bool = false
var сheck *bool = &checking_if_new_processing_has_started

var star_proc int
var end_proc int

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/index.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "index", nil)
}

func current(w http.ResponseWriter, r *http.Request) {

	if !*сheck {
		tmpl, _ := template.ParseFiles("web/current_start.html", "web/header.html", "web/footer.html")
		tmpl.ExecuteTemplate(w, "current_start", nil)
		return
	}

	tmpl, _ := template.ParseFiles("web/current.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "current", nil)
}

func current_graph(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/current_graph.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "current_graph", nil)
}

func success(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/success.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "success", nil)
}

func lib(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/lib.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "lib", nil)
}

func history(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/history.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "history", nil)
}

func base(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/base.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "base", nil)
}

func favourites(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/favourites.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "favourites", nil)
}

func grouping_mode(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/grouping_mode.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "grouping_mode", nil)
}

func grouping_mode_done(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/grouping_mode_done.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "grouping_mode_done", nil)
}

func grouping_mode_done_current(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("web/grouping_mode_done_current.html", "web/header.html", "web/footer.html")
	tmpl.ExecuteTemplate(w, "grouping_mode_done_current", nil)
}



func submit(w http.ResponseWriter, r *http.Request) {
	start := r.FormValue("FormStart")
	end := r.FormValue("FormEnd")
	startConv, err := strconv.Atoi(start)
	star_proc = startConv
	if err != nil {
		fmt.Println(err)

	}
	endConv, err := strconv.Atoi(end)
	end_proc = endConv
	if err != nil {
		fmt.Println(err)

	}
	proc.StandartProc(float64(startConv), float64(endConv))
	*сheck = true
	http.Redirect(w, r, "/current", http.StatusSeeOther)
}

func upload(w http.ResponseWriter, r *http.Request) {
	*сheck = false
	os.Remove("web/temp/")
	os.MkdirAll("web/temp/files/", os.ModePerm)
	os.MkdirAll("web/temp/charts/", os.ModePerm)
	os.MkdirAll("web/temp/bar/", os.ModePerm)

	http.Redirect(w, r, "/success", http.StatusSeeOther)
	
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 40)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("web/temp/files/file.xlsx", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	} else {
		fmt.Println("Incorrect method of interacting with the form")
	}

}

func read_from_group_list(w http.ResponseWriter, r *http.Request) {
	str_vals := make([]string, 0)
	i := 1
	for {
		form_val := r.FormValue(fmt.Sprintf("%d", i))
		if form_val == "" {
			break
		}
		str_vals = append(str_vals, form_val)
		i++
	}
	int_vals := make([][]int, i)
	
	for i, str_val := range str_vals {
		str_arr := strings.Split(str_val, ",")
		for _, str_el := range str_arr {
			int_val, err := strconv.Atoi(str_el)
			if err != nil {
				fmt.Println(err)
			}
			int_vals[i] = append(int_vals[i], int_val)
		}

	}
	proc.GroupModeProc(int_vals[:len(int_vals)-1], float64(star_proc), float64(end_proc))
	http.Redirect(w, r, "/grouping_mode_done", http.StatusSeeOther)

}

func handleRequest() {
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("./web/"))))
	http.HandleFunc("/", current)
	http.HandleFunc("/index", index)
	http.HandleFunc("/lib", lib)
	http.HandleFunc("/history", history)
	http.HandleFunc("/base", base)
	http.HandleFunc("/favourites", favourites)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/submit", submit)
	http.HandleFunc("/current", current)
	http.HandleFunc("/grouping_mode", grouping_mode)
	http.HandleFunc("/current_graph", current_graph)
	http.HandleFunc("/grouping_mode_done", grouping_mode_done)
	http.HandleFunc("/grouping_mode_done_current", grouping_mode_done_current)
	http.HandleFunc("/read_from_group_list", read_from_group_list)
	http.HandleFunc("/success", success)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleRequest()
}
