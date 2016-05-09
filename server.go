// server.go
//
// REST APIs with Go and MySql.
//
// Usage:
//
//   # run go server in the background
//   $ go run server.go 

package main

import (
	"fmt"
    	"io/ioutil"
	"strconv"
	"log"
	"os"
    	"strings"
	"net/http"
    	"encoding/json"
    	"database/sql"
    	_ "github.com/go-sql-driver/mysql"
)


type EMP  struct {
    Id int
    Name string
}

//Handle all requests
func Handler(response http.ResponseWriter, request *http.Request){
    response.Header().Set("Content-type", "text/html")
    webpage, err := ioutil.ReadFile("index.html")
    if err != nil {
    http.Error(response, fmt.Sprintf("home.html file error %v", err), 500)
    }
    fmt.Fprint(response, string(webpage));
}

// Respond to URLs of the form /generic/...
func APIHandler(response http.ResponseWriter, request *http.Request){

	db_url := "admin:admin" + "@tcp(" + os.Getenv("DB_URL") + ")/" + "EMP"
	
    //Connect to database
    db, e := sql.Open("mysql", db_url)
     if( e != nil){
      fmt.Print(e)
     }

    //set mime type to JSON
    response.Header().Set("Content-type", "application/json")
    

	err := request.ParseForm()
	if err != nil {
		http.Error(response, fmt.Sprintf("error parsing url %v", err), 500)
	}

    //can't define dynamic slice in golang
    var result = make([]string,1000)

    switch request.Method {
        case "GET":
            st, err := db.Prepare("select * from employees")
             if err != nil{
              fmt.Print( err );
             }
             rows, err := st.Query()
             if err != nil {
              fmt.Print( err )
             }
             i := 0
             for rows.Next() {
              var name string
              var id int
              err = rows.Scan( &id, &name )
              panda := &EMP{Id: id,Name:name}
                b, err := json.Marshal(panda)
                if err != nil {
                    fmt.Println(err)
                    return
                }
              result[i] = fmt.Sprintf("%s", string(b))
              i++
             }
            result = result[:i]

        case "POST":
            name := request.PostFormValue("name")
	    id := request.PostFormValue("id")
            st, err := db.Prepare("INSERT INTO employees(name,id) VALUES(?,?)")
             if err != nil{
              fmt.Print( err );
             }
             res, err := st.Exec(name,id)
             if err != nil {
              fmt.Print( err )
             }

             if res!=nil{
                 result[0] = "true"
             }
            result = result[:1]

        case "PUT":
            name := request.PostFormValue("name")
            id := request.PostFormValue("id")

            st, err := db.Prepare("UPDATE employees SET name=? WHERE id=?")
             if err != nil{
              fmt.Print( err );
             }
             res, err := st.Exec(name,id)
             if err != nil {
              fmt.Print( err )
             }

             if res!=nil{
                 result[0] = "true"
             }
            result = result[:1]
        case "DELETE":
            id := strings.Replace(request.URL.Path,"/api/id=","",-1)
//		id := request.PostFormValue("id")
            st, err := db.Prepare("DELETE FROM employees WHERE id=?")
             if err != nil{
              fmt.Print( err );
             }
             res, err := st.Exec(id)
             if err != nil {
              fmt.Print( err )
             }

             if res!=nil{
                 result[0] = "true"
             }
            result = result[:1]

        default:
    }
    
    json, err := json.Marshal(result)
    if err != nil {
        fmt.Println(err)
        return
    }

	// Send the text diagnostics to the client.
    fmt.Fprintf(response,"%v",string(json))
	//fmt.Fprintf(response, " request.URL.Path   '%v'\n", request.Method)
    db.Close()
}


func main(){
	port := 1234
    var err string
	portstring := strconv.Itoa(port)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.HandlerFunc( APIHandler ))
	mux.Handle("/", http.HandlerFunc( Handler ))

	// Start listing on a given port with these routes on this server.
	log.Print("Listening on port " + portstring + " ... ")
	errs := http.ListenAndServe(":" + portstring, mux)
	if errs != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
