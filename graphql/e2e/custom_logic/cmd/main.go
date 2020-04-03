package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

func handleCustomRequest(r *http.Request, expectedMethod, resKey string) ([]byte, error) {
	if r.Method != expectedMethod {
		return nil, fmt.Errorf(`{ "errors": [{"message": "Invalid HTTP method: %s"}] }`,
			r.Method)
	}

	if !strings.HasSuffix(r.URL.String(), "/0x123?name=Author&num=10") {
		return nil, fmt.Errorf(`{ "errors": [{"message": "Invalid URL: %s"}] }`, r.URL.String())
	}

	resTemplate := `{
		"%s": [
			{
				"id": "0x3",
				"name": "Star Wars",
				"director": [
					{
						"id": "0x4",
						"name": "George Lucas"
					}
				]
			},
			{
				"id": "0x5",
				"name": "Star Trek",
				"director": [
					{
						"id": "0x6",
						"name": "J.J. Abrams"
					}
				]
			}
		]
	}`

	return []byte(fmt.Sprintf(resTemplate, resKey)), nil
}

func getFavMoviesHandler(w http.ResponseWriter, r *http.Request) {
	b, err := handleCustomRequest(r, http.MethodGet, "myFavoriteMovies")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

func postFavMoviesHandler(w http.ResponseWriter, r *http.Request) {
	b, err := handleCustomRequest(r, http.MethodPost, "myFavoriteMoviesPost")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

func verifyHeadersHandler(w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	expectedKeys := []string{"Accept-Encoding", "User-Agent", "X-App-Token", "X-User-Id"}

	actualKeys := make([]string, 0, len(headers))
	for k := range headers {
		actualKeys = append(actualKeys, k)
	}
	sort.Strings(expectedKeys)
	sort.Strings(actualKeys)
	if !reflect.DeepEqual(expectedKeys, actualKeys) {
		fmt.Fprint(w, `{"errors": [ { "message": "Expected headers not same as actual." } ]}`)
		return
	}

	appToken := r.Header.Get("X-App-Token")
	if appToken != "app-token" {
		fmt.Fprintf(w, `{"errors": [ { "message": "Unexpected value for X-App-Token header: %s." } ]}`, appToken)
		return
	}
	userId := r.Header.Get("X-User-Id")
	if userId != "123" {
		fmt.Fprintf(w, `{"errors": [ { "message": "Unexpected value for X-User-Id header: %s." } ]}`, userId)
		return
	}
	fmt.Fprintf(w, `{"verifyHeaders":[{"id":"0x3","name":"Star Wars"}]}`)
}

func emptyQuerySchema(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
	{
	"data": {
		"__schema": {
		  "queryType": {
			"name": "Query"
		  },
		  "mutationType": null,
		  "subscriptionType": null,
		  "types": [
			{
			  "kind": "OBJECT",
			  "name": "Query",
			  "fields": []
			}]
		  }
	   }
	}
	`)
}

func invalidArgument(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
	{
	"data": {
		"__schema": {
		  "queryType": {
			"name": "Query"
		  },
		  "mutationType": null,
		  "subscriptionType": null,
		  "types": [
			{
			  "kind": "OBJECT",
			  "name": "Query",
			  "fields": [
				{
					"name": "country",
					"args": [
					  {
						"name": "no_code",
						"type": {
						  "kind": "NON_NULL",
						  "name": null,
						  "ofType": {
							"kind": "SCALAR",
							"name": "ID",
							"ofType": null
						  }
						},
						"defaultValue": null
					  }
					],
					"type": {
					  "kind": "OBJECT",
					  "name": "Country",
					  "ofType": null
					},
					"isDeprecated": false,
					"deprecationReason": null
				  }
			  ]
			}]
		  }
	   }
	}
	`)
}

func invalidType(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
	{
	"data": {
		"__schema": {
		  "queryType": {
			"name": "Query"
		  },
		  "mutationType": null,
		  "subscriptionType": null,
		  "types": [
			{
			  "kind": "OBJECT",
			  "name": "Query",
			  "fields": [
				{
					"name": "country",
					"args": [
					  {
						"name": "code",
						"type": {
						  "kind": "NON_NULL",
						  "name": null,
						  "ofType": {
							"kind": "SCALAR",
							"name": "Int",
							"ofType": null
						  }
						},
						"defaultValue": null
					  }
					],
					"type": {
					  "kind": "OBJECT",
					  "name": "Country",
					  "ofType": null
					},
					"isDeprecated": false,
					"deprecationReason": null
				  }
			  ]
			}]
		  }
	   }
	}
	`)
}

func validCountryResponse(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	if strings.Contains(string(body), "__schema") {
		fmt.Fprintf(w, `
	{
	"data": {
		"__schema": {
		  "queryType": {
			"name": "Query"
		  },
		  "mutationType": null,
		  "subscriptionType": null,
		  "types": [
			{
			  "kind": "OBJECT",
			  "name": "Query",
			  "fields": [
				{
					"name": "country",
					"args": [
					  {
						"name": "code",
						"type": {
						  "kind": "NON_NULL",
						  "name": null,
						  "ofType": {
							"kind": "SCALAR",
							"name": "ID",
							"ofType": null
						  }
						},
						"defaultValue": null
					  }
					],
					"type": {
					  "kind": "OBJECT",
					  "name": "Country",
					  "ofType": null
					},
					"isDeprecated": false,
					"deprecationReason": null
				  }
			  ]
			}]
		  }
	   }
	}
	`)
		return
	}

	fmt.Fprintf(w, `
	{
		"data": {
		  "country": {
			"name": "Burundi",
			"code": "BI"
		  }
		}
	  }`)
}
func main() {

	http.HandleFunc("/favMovies/", getFavMoviesHandler)
	http.HandleFunc("/favMoviesPost/", postFavMoviesHandler)
	http.HandleFunc("/verifyHeaders", verifyHeadersHandler)
	http.HandleFunc("/noquery", emptyQuerySchema)
	http.HandleFunc("/invalidargument", invalidArgument)
	http.HandleFunc("/invalidtype", invalidType)
	http.HandleFunc("/validcountry", validCountryResponse)
	fmt.Println("Listening on port 8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
