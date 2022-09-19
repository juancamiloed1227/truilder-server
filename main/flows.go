package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/machinebox/graphql"
)

type flowResource struct{}

type Python struct {
	Code string
}

type Flow struct {
	Title   string
	Content string
}

func (rs flowResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rs.List)            // GET /flows - Read a list of flows.
	r.Post("/", rs.Create)         // POST /flows - Create a new flow.
	r.Post("/execute", rs.Execute) // POST /flows/execute - Execute a flow

	r.Route("/{id}", func(r chi.Router) {
		r.Use(PostCtx)
		r.Get("/", rs.Get)       // GET /flows/{id} - Read a single flow by id
		r.Put("/", rs.Update)    // PUT /flows/{id} - Update a single flow by id
		r.Delete("/", rs.Delete) // DELETE /flows/{id} - Delete a single flow by id
	})

	return r
}

// Request Handler - GET /flows - Read a list of flows.
func (rs flowResource) List(w http.ResponseWriter, r *http.Request) {
	graphqlClient := graphql.NewClient("")
	graphqlRequest := graphql.NewRequest(`
        query {
            queryFlow {
                id
				title
                content
            }
        }
    `)

	var graphqlResponse interface{}
	if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
	}

	jsonStr, err := json.Marshal(graphqlResponse)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		w.Write(jsonStr)
	}
}

// Request Handler - POST /flows - Create a new flow.
func (rs flowResource) Create(w http.ResponseWriter, r *http.Request) {
	var f Flow

	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graphqlClient := graphql.NewClient("")
	query := fmt.Sprintf(`
		mutation addFlow {
				addFlow(input: {title: "%s", content: "%s"}) {
				numUids
			}
		}
	`, f.Title, f.Content)

	graphqlRequest := graphql.NewRequest(query)
	var graphqlResponse interface{}
	if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
	}

	jsonStr, err := json.Marshal(graphqlResponse)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		w.Write(jsonStr)
	}
}

// Request Handler - POST /flows/execute - Execute a flow.
func (rs flowResource) Execute(w http.ResponseWriter, r *http.Request) {
	var p Python

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := exec.Command("python")
	re := strings.NewReader(p.Code)
	var o bytes.Buffer

	cmd.Stdin = re
	cmd.Stdout = &o

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	resp := make(map[string]string)
	resp["response"] = o.String()

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.Write(jsonResp)
}

func PostCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "id", chi.URLParam(r, "id"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Request Handler - GET /flows/{id} - Read a single flow by :id
func (rs flowResource) Get(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)

	graphqlClient := graphql.NewClient("")
	query := fmt.Sprintf(`
		query {
			getFlow(id: "%s") {
				id
				title
				content
			}
		}
	`, id)

	graphqlRequest := graphql.NewRequest(query)
	var graphqlResponse interface{}
	if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
	}

	jsonStr, err := json.Marshal(graphqlResponse)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		w.Write(jsonStr)
	}

}

// Request Handler - PUT /flows/{id} - Update a single flow by :id
func (rs flowResource) Update(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)

	var f Flow

	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graphqlClient := graphql.NewClient("")
	query := fmt.Sprintf(`
		mutation updateFlow {
			updateFlow(input: {filter: {id: "%s"}, set: {content: "%s", title: "%s"}}) {
				numUids
			}
		}
	`, id, f.Content, f.Title)

	graphqlRequest := graphql.NewRequest(query)
	var graphqlResponse interface{}
	if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
	}

	jsonStr, err := json.Marshal(graphqlResponse)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		w.Write(jsonStr)
	}

}

// Request Handler - DELETE /flows/{id} - Delete a single flow by :id
func (rs flowResource) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)

	graphqlClient := graphql.NewClient("")
	query := fmt.Sprintf(`
		mutation deleteFlow {
			deleteFlow(filter: {id: "%s"}) {
				flow {
					id
				}
			}
		}
	`, id)

	graphqlRequest := graphql.NewRequest(query)
	var graphqlResponse interface{}
	if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
	}

	jsonStr, err := json.Marshal(graphqlResponse)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		w.Write(jsonStr)
	}

}
