package application

import (
	"dgb/meter.oil.readings/internal/configuration"
	"dgb/meter.oil.readings/internal/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const route = "/api"
const readingRoute = "/oil"
const SKIP = "skip"
const TAKE = "take"
const SORT = "sort"
const FILTER = "filter"
const ACCESS_CLAIM = "access_as_user"

type ReadingApi struct {
	response   *Response
	config     configuration.Configuration
	repository *database.Repository
	middleware *Middleware
}

func (api *ReadingApi) get(w http.ResponseWriter, r *http.Request) {
	skip, _ := strconv.Atoi(r.FormValue(SKIP))
	take, _ := strconv.Atoi(r.FormValue(TAKE))
	sort := r.FormValue(SORT)
	filter := r.FormValue(FILTER)

	result := api.repository.GetAll(database.PageParams{Skip: skip, Take: take, SortDirection: sort, Filter: filter})

	if result == nil {
		api.response.NotFound(ResponseParams{W: w})
		return
	}

	api.response.Ok(ResponseParams{w, result})
}

func (api *ReadingApi) getTotalForYear(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(mux.Vars(r)["year"])

	result := api.repository.GetTotalForYear(year)

	if result == nil {
		api.response.NotFound(ResponseParams{W: w})
		return
	}

	total := getTotalForYear(result)

	api.response.Ok(ResponseParams{W: w, Result: bson.M{"total": total}})
}

func (api *ReadingApi) count(w http.ResponseWriter, r *http.Request) {
	filter := r.FormValue(FILTER)

	result := api.repository.Count(database.PageParams{Filter: filter})
	api.response.Ok(ResponseParams{w, result})
}

func (api *ReadingApi) getById(w http.ResponseWriter, r *http.Request) {
	result := api.repository.GetSingle(mux.Vars(r)["id"])

	if result == nil {
		api.response.NotFound(ResponseParams{W: w})
		return
	}

	api.response.Ok(ResponseParams{w, result})
}

func (api *ReadingApi) update(w http.ResponseWriter, r *http.Request) {

	var reading primitive.M

	if err := json.NewDecoder(r.Body).Decode(&reading); err != nil {
		api.response.BadRequest(ResponseParams{W: w})
		return
	}

	err := api.repository.Update(mux.Vars(r)["id"], reading)

	if err != nil {
		api.response.ServerError(ResponseParams{W: w})
	}

	api.response.Ok(ResponseParams{W: w})
}

func (api *ReadingApi) create(w http.ResponseWriter, r *http.Request) {
	var reading primitive.M

	if err := json.NewDecoder(r.Body).Decode(&reading); err != nil {
		api.response.BadRequest(ResponseParams{W: w})
		return
	}

	id, err := api.repository.Insert(reading)

	if err != nil {
		api.response.ServerError(ResponseParams{W: w})
	}

	api.response.Created(ResponseParams{w, &Created{
		Url: fmt.Sprintf("%v/%v", route, id),
	}})
}

func (api *ReadingApi) delete(w http.ResponseWriter, r *http.Request) {

	deletedCount, err := api.repository.Delete(mux.Vars(r)["id"])

	if deletedCount <= 0 {
		api.response.NotFound(ResponseParams{W: w})
	}

	if err != nil {
		api.response.ServerError(ResponseParams{W: w})
	}

	api.response.Ok(ResponseParams{W: w})
}

func (api *ReadingApi) HandleRequests() {

	myRouter := mux.NewRouter().StrictSlash(true)
	subRoute := myRouter.PathPrefix(route).Subrouter()

	subRoute.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		api.response.Write(w, 200, "OK")
	}).Methods(http.MethodGet)

	subRoute.
		Path(api.getReadingRoute("count")).
		Queries("filter", "{filter}").
		Methods(http.MethodGet, http.MethodOptions).
		HandlerFunc(api.middleware.Options(api.middleware.Authorize(api.count, ACCESS_CLAIM)))

	subRoute.HandleFunc(api.getReadingRoute("{id:[0-9a-zA\\-]+}"), api.middleware.Options(api.middleware.Authorize(api.getById, ACCESS_CLAIM))).Methods(http.MethodGet, http.MethodOptions)
	subRoute.HandleFunc(api.getReadingRoute("{id:[0-9a-zA\\-]+}"), api.middleware.Options(api.middleware.Authorize(api.update, ACCESS_CLAIM))).Methods(http.MethodPut, http.MethodOptions)
	subRoute.HandleFunc(api.getReadingRoute("{id:[0-9a-zA\\-]+}"), api.middleware.Options(api.middleware.Authorize(api.delete, ACCESS_CLAIM))).Methods(http.MethodDelete, http.MethodOptions)
	subRoute.HandleFunc(api.getReadingRoute("{id:[0-9a-zA\\-]+}"), api.middleware.Options(api.middleware.Authorize(api.create, ACCESS_CLAIM))).Methods(http.MethodPost, http.MethodOptions)

	subRoute.
		Path(api.getReadingRoute("/")).
		Queries("skip", "{skip:[0-9]+}", "take", "{take:[0-9]+}", "sort", "{sort}", "filter", "{filter}").
		Methods(http.MethodGet, http.MethodOptions).
		HandlerFunc(api.middleware.Options(api.middleware.Authorize(api.get, ACCESS_CLAIM)))

	subRoute.
		Path(api.getReadingRoute("{year:[0-9]+}/total")).
		Methods(http.MethodGet, http.MethodOptions).
		HandlerFunc(api.middleware.Options(api.middleware.Authorize(api.getTotalForYear, ACCESS_CLAIM)))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", api.config.HTTP_PORT), subRoute))
}

func (api *ReadingApi) getReadingRoute(subRoute string) string {
	return fmt.Sprintf("%s/%s", readingRoute, subRoute)
}

func NewApi(response *Response, configuration configuration.Configuration, repository *database.Repository, middleware *Middleware) *ReadingApi {
	return &ReadingApi{
		response,
		configuration,
		repository,
		middleware,
	}
}
