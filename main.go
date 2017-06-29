package main

import (
	"database/sql"  //manerar querys
	"fmt"           //formatea, imprimir por consola
	"html/template" //maneja plantillas html
	"log"           // imprime por consola
	"net/http"      //manejar el servidor http
	"strconv"       //conversion de string

	_ "github.com/lib/pq" //libreria de postgres
)

var db *sql.DB
var tpl *template.Template

func init() {

	var err error
	db, err = sql.Open("postgres", "postgres://andres:kelevra@localhost/club?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("Connection succesful")

	tpl = template.Must(template.ParseGlob("templates/*.gohtml")) //inciar todas las plantillas
}

type Team struct {
	Tid   int
	Tname string
	Tdate string
}

type Player struct {
	Pid       int
	Bdate     string
	Pname     string
	Pteam     int
	Pnicnkame *string
	Pposition string
}

func main() {
	//maneja las paginas con funciones
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("pub")))) //Sirve el directorio "PUB" para encontrar cualquier archivo ahí
	http.HandleFunc("/", index)
	http.HandleFunc("/Club", club)
	http.HandleFunc("/Players", players)
	http.HandleFunc("/NewSigning", signing)
	http.HandleFunc("/NewSigning/signed", signed)
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "index.gohtml", nil)
	if err != nil {
		log.Println(err)
	}
}

func club(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT * FROM team")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close() // no meter mas datos en rows

	clbs := make([]Team, 0)
	for rows.Next() {
		clb := Team{}
		err := rows.Scan(&clb.Tid, &clb.Tname, &clb.Tdate) // algun error al escanear la fila
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		clbs = append(clbs, clb)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tpl.ExecuteTemplate(w, "Club.gohtml", clbs)

}

func players(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT * FROM player")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	plys := make([]Player, 0)
	for rows.Next() {
		ply := Player{}
		err := rows.Scan(&ply.Pid, &ply.Bdate, &ply.Pname, &ply.Pteam, &ply.Pnicnkame, &ply.Pposition)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		plys = append(plys, ply)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tpl.ExecuteTemplate(w, "Players.gohtml", plys)

}

func signing(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM team")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	clbs := make([]Team, 0)
	for rows.Next() {
		clb := Team{}
		err := rows.Scan(&clb.Tid, &clb.Tname, &clb.Tdate)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		clbs = append(clbs, clb)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	tpl.ExecuteTemplate(w, "NewSigning.gohtml", clbs)

}

func signed(w http.ResponseWriter, r *http.Request) {
	ply := Player{}
	id := r.FormValue("iden")
	ply.Pname = r.FormValue("Name")
	nicnkame := r.FormValue("pnick")
	ply.Pposition = r.FormValue("position")
	ply.Bdate = r.FormValue("bd")
	team := r.FormValue("pteam")

	i, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, http.StatusText(406)+"Ingrese una identificación valida", http.StatusNotAcceptable)
		return
	}
	ply.Pid = int(i)

	t, err := strconv.ParseInt(team, 10, 32)
	if err != nil {
		http.Error(w, http.StatusText(406)+"Ingrese un equipo valido", http.StatusNotAcceptable)
		return
	}
	ply.Pteam = int(t)
	ply.Pnicnkame = &nicnkame

	_, err = db.Exec("INSERT INTO player (player_id,birth_date,player_name,team,nickname,position) VALUES($1,$2,$3,$4,$5,$6)", ply.Pid, ply.Bdate, ply.Pname, ply.Pteam, ply.Pnicnkame, ply.Pposition)
	if err != nil {
		fmt.Print(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "index.gohtml", nil)
}
