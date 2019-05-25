package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2/bson"

	. "video-reproduction-ms/config"
	. "video-reproduction-ms/dao"
	. "video-reproduction-ms/models"

	"github.com/gorilla/mux"
)

var config = Config{}
var dao = VideosDAO{}

const VIDEO_DIR = "."

const BUFSIZE = 1024 * 8

// Parse the configuration file 'config.toml', and establish a connection to DB
func init() {
	config.Read()

	dao.Server = config.Server
	dao.Database = config.Database
	dao.Connect()
}

// GET list of videos
func AllVideosEndPoint(w http.ResponseWriter, r *http.Request) {
	videos, err := dao.FindAllVideos()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, videos)
}

// GET a video by its ID
func FindVideoEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	video, err := dao.FindVideoById(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Movie ID")
		return
	}
	respondWithJson(w, http.StatusOK, video)
}

// POST a new video
func CreateVideoEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var video Video
	if err := json.NewDecoder(r.Body).Decode(&video); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	video.ID = bson.NewObjectId()
	if video.Destination == "" || video.User_ID == 0 || video.Title == "" || video.Category_ID == "" {
		respondWithError(w, http.StatusInternalServerError, "Empty Values")
		return
	}
	if err := dao.InsertVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusCreated, video)
}

// PUT update an existing video
func UpdateVideoEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var video Video
	if err := json.NewDecoder(r.Body).Decode(&video); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}

// DELETE an existing video
func DeleteVideoEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var video Video
	if err := json.NewDecoder(r.Body).Decode(&video); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.DeleteVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func FindCommentsEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	comments, err := dao.FindCommentsByVideoId(params["video_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Action")
	}

	respondWithJson(w, http.StatusOK, comments)
}

func FindVideoByCategoryEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	videos, err := dao.FindVideoByCategory(params["category_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Action")
	}

	respondWithJson(w, http.StatusOK, videos)
}

func CreateCommentEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var comment Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	comment.ID = bson.NewObjectId()
	if err := dao.InsertComment(comment); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusCreated, comment)
}

func CreateCategoryEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var category Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	category.ID = bson.NewObjectId()

	if category.Description == "" || category.Name == "" {
		respondWithError(w, http.StatusInternalServerError, "Empty Values")
		return
	}

	if err := dao.InsertCategory(category); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusCreated, category)
}

func AllCategoriesEndpoint(w http.ResponseWriter, r *http.Request) {
	categories, err := dao.FindAllCategories()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, categories)
}

func FindCategoryEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	category, err := dao.FindCategoryById(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Category ID")
		return
	}
	respondWithJson(w, http.StatusOK, category)
}

func StreamEndpoint(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	video, err := dao.FindVideoById(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Movie ID")
		return
	}
	path := VIDEO_DIR + video.Destination
	//path := video.URL

	file, err := os.Open(path)
	//file, err := os.Open(path)

	if err != nil {
		w.WriteHeader(500)
		return
	}

	defer file.Close()

	fi, err := file.Stat()

	if err != nil {
		w.WriteHeader(500)
		return
	}

	fileSize := int(fi.Size())

	if len(r.Header.Get("Range")) == 0 {

		contentLength := strconv.Itoa(fileSize)
		contentEnd := strconv.Itoa(fileSize - 1)

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", contentLength)
		w.Header().Set("Content-Range", "bytes 0-"+contentEnd+"/"+contentLength)
		w.WriteHeader(200)

		buffer := make([]byte, BUFSIZE)

		for {
			n, err := file.Read(buffer)

			if n == 0 {
				break
			}

			if err != nil {
				break
			}

			data := buffer[:n]
			w.Write(data)
			w.(http.Flusher).Flush()
		}

	} else {

		rangeParam := strings.Split(r.Header.Get("Range"), "=")[1]
		splitParams := strings.Split(rangeParam, "-")

		// response values

		contentStartValue := 0
		contentStart := strconv.Itoa(contentStartValue)
		contentEndValue := fileSize - 1
		contentEnd := strconv.Itoa(contentEndValue)
		contentSize := strconv.Itoa(fileSize)

		if len(splitParams) > 0 {
			contentStartValue, err = strconv.Atoi(splitParams[0])

			if err != nil {
				contentStartValue = 0
			}

			contentStart = strconv.Itoa(contentStartValue)
		}

		if len(splitParams) > 1 {
			contentEndValue, err = strconv.Atoi(splitParams[1])

			if err != nil {
				contentEndValue = fileSize - 1
			}

			contentEnd = strconv.Itoa(contentEndValue)
		}

		contentLength := strconv.Itoa(contentEndValue - contentStartValue + 1)

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", contentLength)
		w.Header().Set("Content-Range", "bytes "+contentStart+"-"+contentEnd+"/"+contentSize)
		w.WriteHeader(206)

		buffer := make([]byte, BUFSIZE)

		file.Seek(int64(contentStartValue), 0)

		writeBytes := 0

		for {
			n, err := file.Read(buffer)

			writeBytes += n

			if n == 0 {
				break
			}

			if err != nil {
				break
			}

			if writeBytes >= contentEndValue {
				data := buffer[:BUFSIZE-writeBytes+contentEndValue+1]
				w.Write(data)
				w.(http.Flusher).Flush()
				break
			}

			data := buffer[:n]
			w.Write(data)
			w.(http.Flusher).Flush()
		}
	}
}

// Define HTTP request routes
func main() {
	fmt.Println("Starting Server")
	r := mux.NewRouter()
	r.HandleFunc("/videos", AllVideosEndPoint).Methods("GET")
	r.HandleFunc("/videos", CreateVideoEndPoint).Methods("POST") //Created for Testing Purposes
	r.HandleFunc("/videos", UpdateVideoEndPoint).Methods("PUT")  //Creates for Testing Purposes
	r.HandleFunc("/videos", DeleteVideoEndPoint).Methods("DELETE")
	r.HandleFunc("/videos/{id}", FindVideoEndpoint).Methods("GET")
	r.HandleFunc("/videos/{video_id}/comments", FindCommentsEndpoint).Methods("GET")
	r.HandleFunc("/comment", CreateCommentEndpoint).Methods("POST")
	r.HandleFunc("/watch/{id}", StreamEndpoint)
	r.HandleFunc("/videos", UpdateVideoEndPoint).Methods("PUT")
	r.HandleFunc("/categories", CreateCategoryEndpoint).Methods("POST") //Created For Test Purposes
	r.HandleFunc("/categories/{category_id}/videos", FindVideoByCategoryEndpoint).Methods("GET")
	r.HandleFunc("/categories", AllCategoriesEndpoint).Methods("GET")
	r.HandleFunc("/categories/{id}", FindCategoryEndpoint).Methods("GET")

	if err := http.ListenAndServe(":3002", r); err != nil {
		log.Fatal(err)
	}
}