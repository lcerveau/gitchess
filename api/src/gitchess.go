package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	//	"net/http/httputil"
)

const gitToken = "cb8c0a84336c6e5d094f1cc40ffa185147718b65"
const gitGraphQLURL = "https://api.github.com/graphql"

type User struct {
	Login     string
	Name      string
	avatarURL string
}

type UserResponse struct {
	Users []User
	Page  PageInfoResponseComponent
}

type RepositoryResponseComponent struct {
	Name        string
	ID          string
	Description string
	IsFork      bool
	Languages   interface{}
}

type Repository struct {
	Name        string
	ID          string
	Description string
	Languages   []string
	IsFork      bool
}

type PageInfoResponseComponent struct {
	HasNextPage bool
	EndCursor   string
}

type RepositoryResponse struct {
	Data map[string]map[string]map[string]interface{}
}

func main() {
	http.HandleFunc("/", root)
	http.HandleFunc("/repositories", repositories)
	http.HandleFunc("/users", repositories)
	http.ListenAndServe(":8000", nil)
}

func queryToRequest(queryString string) string {
	return `{"query":` + queryString + `}`
}

func contentTypeForRequest(r *http.Request) string {
	var contentType = "html"
	contentType = r.Header.Get("Accept")

	switch contentType {
	case "text/html":
		contentType = "html"
	case "text/xml":
		contentType = "xml"
	case "application/json":
		contentType = "json"
	default:
		c := r.URL.Query().Get("format")
		switch c {
		case "html", "json", "xml":
			contentType = c
		default:
			contentType = "html"
		}
	}
	return contentType
}

// Will get a list of repositories with simple information
func repositories(w http.ResponseWriter, r *http.Request) {

	client := &http.Client{}
	var endCursorID string = ""
	var doStopFetching bool = false
	var httpResponse string = "<html><head></head><body><table>"
	for {
		var qTC string
		if endCursorID == "" {
			qTC = "\"{organization(login:\\\"znly\\\") { repositories(first:20) { nodes { name, id, description, isFork, languages(first:20) {nodes {name}}, refs(first:100, refPrefix:\\\"refs/heads/\\\"){ nodes { name, target {... on Commit{author{name, date} } }}} } pageInfo {hasNextPage, endCursor }}}}\""
		} else {
			qTC = "\"{organization(login:\\\"znly\\\") { repositories(first:20 after:\\\"" + endCursorID + "\\\") { nodes { name, id, description, isFork, languages(first:20){nodes{name }}, refs(first:100, refPrefix:\\\"refs/heads/\\\"){ nodes { name, target { ... on Commit {author{name,date}}} } }} pageInfo {hasNextPage, endCursor }}}}\""
		}
		req, err := http.NewRequest("POST", gitGraphQLURL, bytes.NewBufferString(queryToRequest(qTC)))
		if err != nil {
			panic(err)
		}

		req.Header.Add("Authorization", "Bearer "+gitToken)
		req.Header.Set("Content-Type", "application/json")
		//fmt.Println("ro")
		//_, err := httputil.DumpRequestOut(req, true)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		fmt.Println(string(bodyBytes))
		if resp.Status != "200 OK" {
			httpResponse = httpResponse + "<h1>Error</h1><p>" + string(bodyBytes) + "</p>"
			break
		}

		var jsonResp RepositoryResponse

		err = json.Unmarshal(bodyBytes, &jsonResp)

		if err != nil {
			panic(err)
		}

		for k, v := range jsonResp.Data["organization"]["repositories"] {
			switch k {
			case "nodes":
				p := v.([]interface{})
				for _, b := range p {
					r := b.(map[string]interface{})
					httpResponse = httpResponse + "<tr><td>" + r["name"].(string) + "</td>"

					var t, ok = r["isFork"].(bool)
					if ok == true {
						if t == true {
							httpResponse = httpResponse + "<tr><td>YES</td>"
						} else {
							httpResponse = httpResponse + "<tr><td>NO</td>"
						}
					} else {
						httpResponse = httpResponse + "<tr><td>--</td>"
					}

					_, ok = r["description"].(string)
					if ok == true {
						httpResponse = httpResponse + "<td>" + r["description"].(string) + "</td></tr>"
					} else {
						httpResponse = httpResponse + "<td>--</td></tr>"
					}

					_, ok = r["languages"].(map[string][]map[string]string)
					if ok == true {
						print("hooura")
					}

				}
			case "pageInfo":
				p := v.(map[string]interface{})
				fmt.Println(p)
				if p["hasNextPage"].(bool) == false {
					doStopFetching = true
					break
				} else {
					endCursorID = p["endCursor"].(string)
				}
			}
		}
		if doStopFetching == true {
			break
		}
	}
	httpResponse = httpResponse + "</body>"
	io.WriteString(w, httpResponse)
}

func users(w http.ResponseWriter, r *http.Request) {

}

func root(w http.ResponseWriter, r *http.Request) {
	var contentType = contentTypeForRequest(r)
	switch contentType {
	case "html":
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<html>\n<head>\n</head>\n<body>\n\t<ul>\n\t\t<li><a href=\"/repositories\">Repositories</a></li>\n\t\t<li><a>Users</a></li>\n\t</ul>\n</body>")
	case "xml":
		w.Header().Set("Content-Type", "text/xml")
		io.WriteString(w, "<pages>\n\t<page>/repositories</page>\n\t<page>/users</page>\n</pages>")
	case "json":
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "{\n\tpages:[\"/repositories\", \"/users\"]\n}")
	}

}
