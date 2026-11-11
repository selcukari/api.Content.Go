package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"web-api/config"
	"web-api/models"
	"web-api/pkg"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Get all users
func GetAllContents(w http.ResponseWriter, r *http.Request) {
	var contents []models.Content

	// Retrieving data from the cache
	if val, ok := pkg.GlobalCache.Get("getAllContents"); ok {

		// Önbellekten gelen veriyi uygun tipe dönüştürüp istemciye dönüyoruz
		if cachedContents, typeOk := val.([]models.Content); typeOk {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cachedContents)
			return
		}
	}

	client, err := config.ConnectToMongoDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.Background())

	collection := client.Database(pkg.GetMongoDBDatabaseName()).Collection("Contents")

	findOptions := options.Find()
	findOptions.SetLimit(50)
	findOptions.SetSort(bson.D{{"UpdatedDate", -1}})

	cursor, err := collection.Find(context.Background(), bson.D{}, findOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var content models.Content
		if err := cursor.Decode(&content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		uuidValue, err := uuid.FromBytes(content.ID.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		content.Id = uuidValue.String()
		contents = append(contents, content)
	}
	// Veriyi cache'e kaydet
	pkg.GlobalCache.Set("getAllContents", contents, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	fmt.Println("cahce getAllContents çalıştı2")

	json.NewEncoder(w).Encode(contents)
}

// Get a user by ID
func GetContentID(w http.ResponseWriter, r *http.Request) {
	client, err := config.ConnectToMongoDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.Background())

	id := strings.TrimPrefix(r.URL.Path, "/contents/")

	// String değerini UUID'ye dönüştür
	uuidValue, err := uuid.Parse(id)

	if err != nil {
		http.Error(w, "Geçersiz UUID", http.StatusBadRequest)
		return
	}

	// UUID'yi BSON Binary subtype 4 biçimine dönüştür
	uuidBinary := primitive.Binary{
		Subtype: 4,
		Data:    uuidValue[:],
	}

	if val, ok := pkg.GlobalCache.Get("getContentID_" + uuidValue.String()); ok {
		// Önbellekten gelen veriyi uygun tipe dönüştürüp istemciye dönüyoruz
		if cachedContent, typeOk := val.(models.Content); typeOk {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cachedContent)
			return
		}
	}

	collection := client.Database(pkg.GetMongoDBDatabaseName()).Collection("Contents")
	var content models.Content
	if err := collection.FindOne(context.Background(), bson.M{"_id": uuidBinary}).Decode(&content); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	getUuidValue, err := uuid.FromBytes(content.ID.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	content.Id = getUuidValue.String()

	pkg.GlobalCache.Set("getContentID_"+uuidValue.String(), content, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(content)
}
