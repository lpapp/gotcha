package main

import (
    "fmt"
    "html/template"
    "log"
    "net"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "sort"
)

var freeOwner = "Free"

var personIpsMap = map[string][]string{
    "Laszlo": []string{"10.104.68.90", "10.104.68.60"},
    "Adam":   []string{"192.168.3.239", "10.104.68.36"},
    "Nikesh": []string{"192.168.3.220"},
    "Hubert": []string{"192.168.3.229"},
    "Michael": []string{"10.104.68.77"},
}

type Switches struct {
	Switches []Switch `json:"switches"`
}

type Switch struct {
        Name string         `json:"name"`
        Description string  `json:"description"`
        Owner string        `json:"owner"`
    }

var switches Switches

func add(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
    } else {
        r.ParseForm()
        defer http.Redirect(w, r, "/", http.StatusSeeOther)
		for _, s := range switches.Switches {
			if s.Name == r.Form["Name"][0] {
                fmt.Println("Already added")
                return
            }
        }
        newSwitch := Switch {
            Name: r.Form["Name"][0],
            Description: r.Form["Description"][0],
            Owner: freeOwner,
        }
        switches.Switches = append(switches.Switches, newSwitch)
        sort.Slice(switches.Switches, func(i, j int) bool {
            return switches.Switches[i].Name < switches.Switches[j].Name
        })
        file, _ := json.MarshalIndent(switches, "", " ")
        _ = ioutil.WriteFile("gotcha.json", file, 0644)
    }
}

func del(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
    } else {
        r.ParseForm()
		for i, s := range switches.Switches {
			if s.Name == r.Form["Name"][0] {
                switches.Switches = append(switches.Switches[:i], switches.Switches[i+1:]...)
                break
            }
		}
        file, _ := json.MarshalIndent(switches, "", " ")
        _ = ioutil.WriteFile("gotcha.json", file, 0644)
		http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

func acquire(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
    } else {
        r.ParseForm()
		for i, s := range switches.Switches {
			if s.Name == r.Form["Name"][0] {
                if s.Owner == freeOwner {
                    rip, _, _ := net.SplitHostPort(r.RemoteAddr)
                    fmt.Println("TEST IP: " + rip)
                    for k, v := range personIpsMap {
                        for _, oip := range v {
                            if oip == rip {
                                switches.Switches[i].Owner = k
                                // Note: some goto would be nicer
                                break
                            }
                        }
                    }
                } else {
                    fmt.Println("Owned by someone")
                }
                break
            }
		}
        file, _ := json.MarshalIndent(switches, "", " ")
        _ = ioutil.WriteFile("gotcha.json", file, 0644)
		http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

func release(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
    } else {
        r.ParseForm()
		for i, s := range switches.Switches {
			if s.Name == r.Form["Name"][0] {
                rip, _, _ := net.SplitHostPort(r.RemoteAddr)
                for k, v := range personIpsMap {
                    for _, oip := range v {
                        if oip == rip {
                            if k == switches.Switches[i].Owner {
                                switches.Switches[i].Owner = freeOwner
                            } else {
                                fmt.Println("Not owned by you!")
                            }
                            // Note: some goto would be nicer
                            break
                        }
                    }
                }
                break
            }
		}
        file, _ := json.MarshalIndent(switches, "", " ")
        _ = ioutil.WriteFile("gotcha.json", file, 0644)
		http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

func list(w http.ResponseWriter, r *http.Request) {
    fmt.Println("method:", r.Method) //get request method
    if r.Method == "GET" {
        file, _ := ioutil.ReadFile("gotcha.json")
        _ = json.Unmarshal([]byte(file), &switches)

        tmp := template.New("lpapp example")
        t, err := tmp.Parse(`
        <html>
            <head>
            <title></title>
            </head>
            <body>
				<table>
				  <tr>
					<th>Name</th>
					<th>Owner</th> 
					<th>Description</th>
				  </tr>
                {{range .Switches}}
				  <tr>
					<td>{{.Name}}</td>
					<td>{{.Owner}}</td> 
					<td>
					{{if .Description}}
                    	{{.Description}}
					{{else}}
						-
					{{end}}
                    </td>
				  </tr>
                {{end}}
				</table>
                <br/><br/><br/>
                <form action="/acquire" method="post">
                    Name:<input type="text" name="Name"><input type="submit" value="Acquire">
                </form>
                <form action="/release" method="post">
                    Name:<input type="text" name="Name"><input type="submit" value="Release">
                </form>
                <br/><br/><br/>
                <form action="/add" method="post">
                    Name:<input type="text" name="Name">    Description:<input type="text" name="Description"><input type="submit" value="Add">
                </form>
                <form action="/del" method="post">
                    Name:<input type="text" name="Name"><input type="submit" value="Delete">
                </form>
            </body>
        </html>`)
		if err != nil {
			fmt.Println(err)
		}
        t.Execute(w, switches)
    }
}

func main() {
    http.HandleFunc("/", list)
    http.HandleFunc("/add", add)
    http.HandleFunc("/del", del)
    http.HandleFunc("/acquire", acquire)
    http.HandleFunc("/release", release)
    err := http.ListenAndServe(":9090", nil) // setting listening port
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
