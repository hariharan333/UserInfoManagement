package controllers

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/mail"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/notblessy/go-ingnerd/src/models"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

var db *gorm.DB = models.ConnectDB()

type userRequest struct {
	FullName  string `json:"fullName"`
	Email     string `json:"email"`
	PhoneNo   string `json:"phoneNo"`
	Image     string `json:"image"`
	CreateaAt string `json:"createAt"`
	UpdatedAt string `json:"updatedAt"`
}

type userResponse struct {
	userRequest
	ID int64 `json:"id"`
}

// User name validation function
func fullNameValidation(fullName string) (string, bool) {
	//Checking the user fullname length
	if len(fullName) < 2 || len(fullName) > 30 {
		return "Full name should be more than two characters. note: Full name length range should be two to thirty characters", false
	}
	//Matching the user fullname with regex
	pattern := regexp.MustCompile("^[a-zA-Z]{1,}(?: [a-zA-Z]+){0,2}$")
	matches := pattern.MatchString(fullName)
	if !matches {
		return "Fullname is invalid. Example of fullname is 'abcd xyz'", false
	}
	return "", true
}

// Email validation function
func emailValidation(user models.User) (string, bool) {

	//We need to confirm that emailid is used by someone or not. because email id is unique.
	result := models.User{}
	db.Where("email = ?", user.Email).First(&result)
	if result.Email != "" && result.Id != user.Id {
		return "Email Id is already used by someone. please try with another emailid. Example of emailId is 'abc@example.com'", false
	}

	//Checking the email address formate
	_, err := mail.ParseAddress(user.Email)
	if err != nil {
		return "Email Id is invalid. Example of emailId is 'abc@example.com'", false
	}
	return "", true
}

// Phone number validation function
func phoneValidation(user models.User) (string, bool) {
	result := models.User{}

	//We need to confirm that phone number is used by someone or not. because phone number is unique.
	db.Where("phone_no = ?", user.PhoneNo).First(&result)
	if result.PhoneNo != "" && result.Id != user.Id {
		return "This phone number is already used by someone. please try with another phone number", false
	}

	//Matching the user phone number with regex
	re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
	if !re.MatchString(user.PhoneNo) {
		return "Phone number is invalid", false
	}
	return "", true
}

// Store the user information in user table
func CreateUser(context *gin.Context) {

	user := models.User{}

	//Processing the user request

	//Get the user fullname from form payload
	user.FullName = context.PostForm("fullName")

	//Validate the user fullname
	msg, valid := fullNameValidation(user.FullName)
	if !valid {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": msg})
		return
	}

	//Get the user emailid from form payload
	user.Email = context.PostForm("email")

	//Validate the user emailid
	msg, valid = emailValidation(user)
	if !valid {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": msg})
		return
	}

	//Get the user phone number from form payload
	user.PhoneNo = context.PostForm("phoneNo")

	//Validate the user phone number
	msg, valid = phoneValidation(user)
	if !valid {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": msg})
		return
	}

	// We can store the user image in two ways
	// 1. Using blob data type to store the image in db
	// 2. store the image in s3 bucket and then storing the s3 bucket image url in user table.
	// It's totally depends on usercase. My usecase is very straigt forward so I'm using the blob to store the binary image in table.
	// If you need the second way ,I can do it.

	context.Request.ParseMultipartForm(10 << 20)
	imageFile := make([]byte, 0)
	image, handler, err := context.Request.FormFile("image")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": "Getting error while reading the user image"})
		return
	} else {
		defer image.Close()
		if handler.Size > 50000 {
			context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": "Image size should be below 50kb"})
			return
		}
		imageFile, err = ioutil.ReadAll(image)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": "Error reading uploaded image from stream"})
			return
		}
	}
	encodedImage := base64.StdEncoding.EncodeToString(imageFile)
	user.Image = encodedImage

	//Set the current date and time to createdAt field
	dt := time.Now()
	user.CreateaAt = dt.Format("01-02-2006 15:04:05")
	user.UpdatedAt = ""

	//Insert the user details into the user table
	result := db.Create(&user)
	if result.Error != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "500", "error": "Getting an error while inserting the user data into the user table"})
		return
	}

	//Return the success respone
	var response userResponse
	response.ID = user.Id
	response.FullName = user.FullName
	response.Email = user.Email
	response.PhoneNo = user.PhoneNo
	response.Image = user.Image
	response.CreateaAt = user.CreateaAt
	response.UpdatedAt = user.UpdatedAt
	context.JSON(http.StatusCreated, gin.H{
		"status code": "200",
		"message":     "Successfull get the all users",
		"data":        response,
	})
}

// Get the all users data from the user table
func GetAllUsers(context *gin.Context) {
	var users []models.User

	//Get the all users data from the user table
	err := db.Find(&users)
	if err.Error != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "500", "error": "Getting an error while getting all users data from the user table"})
		return
	}

	//Return the success response
	context.JSON(http.StatusOK, gin.H{
		"status code": "200",
		"message":     "Successfull get the all users",
		"data":        users,
	})

}

// Get the user data from user table based on user id
func GetUser(context *gin.Context) {
	var users []models.User

	//Get the userid path param from the request url
	reqParamId := context.Param("userid")
	userid := cast.ToUint(reqParamId)

// 	//Get the user data from user table based on user id
// 	err := db.First(&user, userid)
// 	if err.Error != nil {
// 		context.JSON(http.StatusBadRequest, gin.H{"status code": "500", "error": "Getting an error while getting a user data from the user table based on id"})
// 		return
// 	}
	//Get the all users data from the user table
	err := db.Find(&users)
	if err.Error != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "500", "error": "Getting an error while getting all users data from the user table"})
		return
	}
	if len(users) < int(userid) {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": "Invalid pagination number. please enter the valid pagination."})
		return
	}
	
	//Return success response
	context.JSON(http.StatusOK, gin.H{
		"status code": "200",
		"message":     "Successfully get the user data",
		"data":        users[userid-1],
	})

}

// Update the user data in user table based on user id
func UpdateUser(context *gin.Context) {

	//Processing the user request

	//Get the userid path param from the request url
	reqParamId := context.Param("userid")
	userid := cast.ToUint(reqParamId)

	//Get the user data from user table based on userid
	user := models.User{}

	userById := db.Where("id = ?", userid).First(&user)
	if userById.Error != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": "Userid is not found"})
		return
	}

	//Get the user fullname from form payload
	user.FullName = context.PostForm("fullName")

	//Validate the user fullname
	msg, valid := fullNameValidation(user.FullName)
	if !valid {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": msg})
		return
	}

	//Get the user emailid from form payload
	user.Email = context.PostForm("email")

	//validate the user emailid
	msg, valid = emailValidation(user)
	if !valid {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": msg})
		return
	}

	//Get the user phone number from form payload
	user.PhoneNo = context.PostForm("phoneNo")

	//Validate the user phone number
	msg, valid = phoneValidation(user)
	if !valid {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": msg})
		return
	}

	// we can store the user image in two ways
	// 1. Using blob data type to store the image in db
	// 2. store the image in s3 bucket and then storing the s3 bucket image url in user table.
	// It's totally depends on usercase. My usecase is very straigt forward so I'm using the blob to store the binary image in table.
	// If you need the second way ,I can do it.

	context.Request.ParseMultipartForm(10 << 20)
	imageFile := make([]byte, 0)
	image, handler, err := context.Request.FormFile("image")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": "Getting error while reading the user image"})
		return
	} else {
		defer image.Close()
		if handler.Size > 50000 {
			context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": "Image size should be below 50kb"})
			return
		}
		imageFile, err = ioutil.ReadAll(image)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": "Error reading uploaded image from stream"})
			return
		}
	}
	encodedImage := base64.StdEncoding.EncodeToString(imageFile)
	user.Image = encodedImage

	//Set the current date and time to updatedAt field
	dt := time.Now()
	user.UpdatedAt = dt.Format("01-02-2006 15:04:05")

	//Update the user information in the user table
	result := db.Save(&user)
	if result.Error != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "500", "error": "Getting an error while updating the user data into the user table"})
		return
	}

	//Return success response
	var response userResponse
	response.ID = user.Id
	response.FullName = user.FullName
	response.Email = user.Email
	response.PhoneNo = user.PhoneNo
	response.Image = user.Image
	response.CreateaAt = user.CreateaAt
	response.UpdatedAt = user.UpdatedAt

	context.JSON(200, gin.H{
		"status code": "200",
		"message":     "Successfully updated",
		"data":        response,
	})
}

func DeleteUser(context *gin.Context) {
	user := models.User{}

	//Get the userid path param from the request url
	reqParamId := context.Param("userid")
	userid := cast.ToUint(reqParamId)

	//Delete the user data from the user table based on the userid
	delete := db.Where("id = ?", userid).Unscoped().Delete(&user)
	if delete.RowsAffected == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"status code": "400", "error": "Userid is not found"})
		return
	}

	//Return success response
	context.JSON(http.StatusOK, gin.H{
		"status code": "200",
		"message":     "Successfully deleted",
		"data":        userid,
	})

}
