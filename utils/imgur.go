package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ImgurResponse struct {
    Success bool `json:"success"`
    Status  int  `json:"status"`
    Data    struct {
        Link       string `json:"link"`
        DeleteHash string `json:"deletehash"`
    } `json:"data"`
}

func UploadToImgur(file *multipart.FileHeader) (*fiber.Map, error) {
    src, err := file.Open()
    if (err != nil) {
        return nil, fiber.NewError(fiber.StatusBadRequest, "Failed to open file")
    }
    defer src.Close()

    // Read file content
    fileBytes, err := io.ReadAll(src)
    if (err != nil) {
        return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to read file")
    }

    // Create request body
    body := new(bytes.Buffer)
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("image", file.Filename)
    if (err != nil) {
        return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create form file")
    }

    if _, err = part.Write(fileBytes); err != nil {
        return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to write file")
    }
    writer.Close()

    // Create request
    req, err := http.NewRequest("POST", "https://api.imgur.com/3/image", body)
    if (err != nil) {
        return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create request")
    }

    // Set headers
    req.Header.Set("Authorization", "Client-ID "+os.Getenv("IMGUR_CLIENT_ID"))
    req.Header.Set("Content-Type", writer.FormDataContentType())

    // Send request
    client := &http.Client{}
    resp, err := client.Do(req)
    if (err != nil) {
        return nil, fiber.NewError(fiber.StatusServiceUnavailable, "Failed to upload to Imgur")
    }
    defer resp.Body.Close()

    // Parse response
    var imgurResp ImgurResponse
    if err := json.NewDecoder(resp.Body).Decode(&imgurResp); err != nil {
        return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to parse Imgur response")
    }

    if !imgurResp.Success {
        return nil, fiber.NewError(fiber.StatusBadGateway, "Imgur upload failed")
    }

    return &fiber.Map{
        "link": imgurResp.Data.Link,
        "deleteHash": imgurResp.Data.DeleteHash, // 
    }, nil
}

func DeleteFromImgur(deleteHash string) error {
    if deleteHash == "" {
        return fiber.NewError(fiber.StatusBadRequest, "Delete hash is required")
    }

    // Create delete request with proper URL
    req, err := http.NewRequest(
        "DELETE", 
        fmt.Sprintf("https://api.imgur.com/3/image/%s", deleteHash), 
        nil,
    )
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to create delete request")
    }

    // Set proper authorization header with Bearer token
    req.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", os.Getenv("IMGUR_CLIENT_ID")))

    // Create client with timeout
    client := &http.Client{
        Timeout: time.Second * 10,
    }

    // Send request
    resp, err := client.Do(req)
    if err != nil {
        return fiber.NewError(fiber.StatusServiceUnavailable, "Failed to connect to Imgur")
    }
    defer resp.Body.Close()

    // Parse response to check success
    var response struct {
        Success bool   `json:"success"`
        Status  int    `json:"status"`
        Data    string `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse Imgur response")
    }

    if !response.Success {
        return fiber.NewError(
            fiber.StatusBadGateway, 
            fmt.Sprintf("Failed to delete image from Imgur. Status: %d", response.Status),
        )
    }

    return nil
}