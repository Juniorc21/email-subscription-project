package main

import (
	"context"
	"fmt"
	"net/http"
	subscribeemails "subscribeemail"

	"github.com/labstack/echo/v4"
	"go.temporal.io/sdk/client"
)

var temporalClient client.Client

type RequestData struct {
	Email string `json:"email"`
}
type ResponseData struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	var err error
	temporalClient, err = client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting the web server on %s\n", subscribeemails.ClientHostPort)

	e := echo.New()
	e.POST("/subscribe", subscribeHandler)
	e.DELETE("/unsubscribe", unsubscribeHandler)
	e.GET("/details", showDetailsHandler)

	e.Logger.Fatal(e.Start(":4000"))
}

func subscribeHandler(ctx echo.Context) error {
	requestData := RequestData{}
	if err := ctx.Bind(&requestData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if requestData.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email")
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:                                       requestData.Email,
		TaskQueue:                                subscribeemails.TaskQueueName,
		WorkflowExecutionErrorWhenAlreadyStarted: true,
	}

	subscription := subscribeemails.EmailDetails{
		EmailAddress:      requestData.Email,
		Message:           "Welcome to the subscription Workflow!",
		SubscriptionCount: 0,
		IsSubscribed:      true,
	}

	_, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, subscribeemails.SubscriptionWorkflow, subscription)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Couldn't sign up user. Please try again.")
	}

	responseData := ResponseData{
		Status:  "success",
		Message: "Sugned up",
	}

	return ctx.JSON(http.StatusCreated, responseData)

}

func showDetailsHandler(ctx echo.Context) error {
	email := ctx.QueryParam("email")

	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "not nil email")
	}

	workflowID := email
	queryType := "GetDetails"

	resp, err := temporalClient.QueryWorkflow(context.Background(), workflowID, "", queryType)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Couldn't query values. Please try again.")
	}

	var result subscribeemails.EmailDetails

	if err := resp.Get(&result); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Couldn't query values. Please try again.")
	}

	return ctx.JSON(http.StatusOK, result)
}

func unsubscribeHandler(ctx echo.Context) error {

	requestData := RequestData{}
	if err := ctx.Bind(&requestData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if requestData.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email")
	}
	workflowID := requestData.Email

	err := temporalClient.CancelWorkflow(context.Background(), workflowID, "")

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Couldn't unsubscribe. Please try again")
	}

	responseData := ResponseData{
		Status:  "success",
		Message: "Unsubscribed.",
	}

	return ctx.JSON(http.StatusAccepted, responseData)
}
