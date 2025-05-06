package handlers

import (
	"context"
	"nitri-meal-backend/database"
	"nitri-meal-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetCommunityPosts retrieves paginated community posts
func GetCommunityPosts(c *fiber.Ctx) error {
    collection := database.GetCollection("community_posts")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Parse query parameters
    page := c.QueryInt("page", 0)
    limit := c.QueryInt("limit", 10)
    skip := page * limit

    // Setup options for sorting and pagination
    opts := options.Find().
        SetSort(bson.D{{Key: "createdAt", Value: -1}}).
        SetSkip(int64(skip)).
        SetLimit(int64(limit))

    // Execute query
    cursor, err := collection.Find(ctx, bson.M{}, opts)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch posts",
        })
    }
    defer cursor.Close(ctx)

    var posts []models.Post
    if err := cursor.All(ctx, &posts); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to decode posts",
        })
    }

    return c.JSON(posts)
}

// CreateCommunityPost creates a new community post
func CreateCommunityPost(c *fiber.Ctx) error {
    // Parse multipart form
    form, err := c.MultipartForm()
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid form data",
        })
    }

    // Validate content
    contentValues := form.Value["content"]
    if len(contentValues) == 0 {
        return c.Status(400).JSON(fiber.Map{
            "error": "Content is required",
        })
    }
    content := contentValues[0]
    if content == "" {
        return c.Status(400).JSON(fiber.Map{
            "error": "Content is required",
        })
    }

    // Validate author data
    authorIDValues := form.Value["author.id"]
    authorNameValues := form.Value["author.name"]
    authorPictureValues := form.Value["author.picture"]
    if len(authorIDValues) == 0 || len(authorNameValues) == 0 || len(authorPictureValues) == 0 {
        return c.Status(400).JSON(fiber.Map{
            "error": "Author information is incomplete",
        })
    }

    // Handle image upload
    var imageData []byte
    var imageType string
    if files := form.File["image"]; len(files) > 0 {
        file := files[0]
        
        // Validate file size
        if file.Size > 5*1024*1024 {
            return c.Status(400).JSON(fiber.Map{
                "error": "Image too large (max 5MB)",
            })
        }

        // Read file
        fileContent, err := file.Open()
        if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": "Failed to process image",
            })
        }
        defer fileContent.Close()

        imageData = make([]byte, file.Size)
        if _, err := fileContent.Read(imageData); err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": "Failed to read image",
            })
        }

        imageType = file.Header.Get("Content-Type")
    }

    // Create new post
    post := &models.Post{
        ID:        primitive.NewObjectID(),
        Content:   content,
        Image:     imageData,
        ImageType: imageType,
        CreatedAt: time.Now(),
        Likes:     0,
        LikedBy:   make([]string, 0),
        Author: models.Author{
            ID:      authorIDValues[0],
            Name:    authorNameValues[0],
            Picture: authorPictureValues[0],
        },
    }

    collection := database.GetCollection("community_posts")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err = collection.InsertOne(ctx, post)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to create post",
        })
    }

    return c.Status(201).JSON(post)
}

// Add this new handler function
func LikePost(c *fiber.Ctx) error {
    // Get post ID from URL
    postID, err := primitive.ObjectIDFromHex(c.Params("postId"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid post ID",
        })
    }

    // Get user ID from request body
    var body struct {
        UserID string `json:"userId"`
    }
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    collection := database.GetCollection("community_posts")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Find the post and check if user already liked it
    var post models.Post
    err = collection.FindOne(ctx, bson.M{"_id": postID}).Decode(&post)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{
            "error": "Post not found",
        })
    }

    // Check if user already liked the post
    isLiked := false
    for _, userID := range post.LikedBy {
        if userID == body.UserID {
            isLiked = true
            break
        }
    }

    var update bson.M
    if isLiked {
        // Unlike: remove user from likedBy and decrease likes count
        update = bson.M{
            "$pull": bson.M{"likedBy": body.UserID},
            "$inc":  bson.M{"likes": -1},
        }
    } else {
        // Like: add user to likedBy and increase likes count
        update = bson.M{
            "$push": bson.M{"likedBy": body.UserID},
            "$inc":  bson.M{"likes": 1},
        }
    }

    result, err := collection.UpdateOne(ctx, bson.M{"_id": postID}, update)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to update likes",
        })
    }

    if result.ModifiedCount == 0 {
        return c.Status(404).JSON(fiber.Map{
            "error": "Post not found",
        })
    }

    return c.Status(200).JSON(fiber.Map{
        "success": true,
        "liked": !isLiked,
    })
}

// DeleteCommunityPost deletes a post if the user owns it
func DeleteCommunityPost(c *fiber.Ctx) error {
    // Get post ID from URL
    postID, err := primitive.ObjectIDFromHex(c.Params("postId"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid post ID",
        })
    }

    // Get user ID from request
    userID := c.Query("userId")
    if userID == "" {
        return c.Status(400).JSON(fiber.Map{
            "error": "User ID is required",
        })
    }

    collection := database.GetCollection("community_posts")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Find and delete the post only if the user is the author
    result, err := collection.DeleteOne(ctx, bson.M{
        "_id": postID,
        "author._id": userID,
    })

    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to delete post",
        })
    }

    if result.DeletedCount == 0 {
        return c.Status(403).JSON(fiber.Map{
            "error": "Not authorized to delete this post",
        })
    }

    return c.Status(200).JSON(fiber.Map{
        "success": true,
        "message": "Post deleted successfully",
    })
}