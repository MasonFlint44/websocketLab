package main

import (
	"github.com/masonflint44/websocketLab/pkg/models"
)

func sendMessages() {
	for {
		m := <-outboundResponses
		pipe(
			clientPipe(
				m.GetClient(),
				hasConn,
			),
			messagePipe(
				m,
				sendResponse,
			),
		)()
	}
}

func processLogoutRequests() {
	for {
		req := <-logoutRequests
		pipe(
			clientPipe(
				req.GetClient(),
				hasConn,
			),
			failPipe(
				clientPipe(
					req.GetClient(),
					hasAuth,
				)(),
				messagePipe(
					models.Message{
						Client: models.Client{
							Conn:   req.GetClient().GetConn(),
							Handle: "Server",
						},
						Body: "Client is not logged in",
					},
					queueMessage,
				),
				closePipe(),
			),
			clientPipe(
				req.GetClient(),
				logout,
			),
			messagePipe(
				models.Message{
					Client: models.Client{
						Conn:   req.GetClient().GetConn(),
						Handle: "Server",
					},
					Body: "Successful logout",
				},
				queueMessage,
			),
		)()
	}
}

func processNewUserRequests() {
	for {
		req := <-newUserRequests
		pipe(
			clientPipe(
				req.GetClient(),
				hasConn,
			),
			messagePipe(
				req.GetMessage(),
			),
			failPipe(
				clientPipe(
					req.GetMessage().GetClient(),
					validHandle,
				)(),
				messagePipe(
					models.Message{
						Client: models.Client{
							Conn:   req.GetClient().GetConn(),
							Handle: "Server",
						},
						Body: "Handle must be less than 32 characters",
					},
					queueMessage,
				),
				closePipe(),
			),
			failPipe(
				clientPipe(
					req.GetMessage().GetClient(),
					validPass,
				)(),
				messagePipe(
					models.Message{
						Client: models.Client{
							Conn:   req.GetClient().GetConn(),
							Handle: "Server",
						},
						Body: "Pass must be between 4 and 8 characters",
					},
					queueMessage,
				),
				closePipe(),
			),
			failPipe(
				clientPipe(
					req.GetMessage().GetClient(),
					uniqueHandle,
				)(),
				messagePipe(
					models.Message{
						Client: models.Client{
							Conn:   req.GetClient().GetConn(),
							Handle: "Server",
						},
						Body: "Handle is already taken",
					},
					queueMessage,
				),
				closePipe(),
			),
			failPipe(
				clientPipe(
					req.GetMessage().GetClient(),
					register,
				)(),
				messagePipe(
					models.Message{
						Client: models.Client{
							Conn:   req.GetClient().GetConn(),
							Handle: "Server",
						},
						Body: "Unable to register new user",
					},
					queueMessage,
				),
				closePipe(),
			),
			messagePipe(
				models.Message{
					Client: models.Client{
						Conn:   req.GetClient().GetConn(),
						Handle: "Server",
					},
					Body: "Welcome " + req.GetMessage().GetClient().GetHandle() + "! Use 'login' to continue.",
				},
				queueMessage,
			),
		)()
	}
}

func processSendRequests() {
	for {
		req := <-sendRequests
		pipe(
			clientPipe(
				req.GetClient(),
				hasConn,
			),
			messagePipe(
				req.GetMessage(),
			),
			failPipe(
				clientPipe(
					req.GetClient(),
					hasAuth,
				)(),
				messagePipe(
					models.Message{
						Client: models.Client{
							Conn:   req.GetClient().GetConn(),
							Handle: "Server",
						},
						Body: "Unauthorized - Please login",
					},
					queueMessage,
				),
				closePipe(),
			),
			messagePipe(
				models.Message{
					Client: models.Client{
						Handle: req.GetClient().GetHandle(),
					},
					Body: req.GetMessage().GetBody(),
				},
				forEachClient(queueMessage)...,
			),
		)()
	}
}

func processLoginRequests() {
	for {
		req := <-loginRequests
		pipe(
			clientPipe(
				req.GetClient(),
				hasConn,
			),
			failPipe(
				clientPipe(
					req.GetClient(),
					hasAuth,
				)(),
				failPipe(
					clientPipe(
						req.GetMessage().GetClient(),
						authorize(
							req.GetClient().GetConn(),
						),
					)(),
					messagePipe(
						models.Message{
							Client: models.Client{
								Conn:   req.GetClient().GetConn(),
								Handle: "Server",
							},
							Body: "Unable to log in with provided credentials",
						},
						queueMessage,
					),
					closePipe()),
				messagePipe(
					models.Message{
						Client: models.Client{
							Conn:   req.GetClient().GetConn(),
							Handle: "Server",
						},
						Body: "Successful login",
					},
					queueMessage,
				),
				closePipe(),
			),
			messagePipe(
				models.Message{
					Client: models.Client{
						Conn:   req.GetClient().GetConn(),
						Handle: "Server",
					},
					Body: "Client is already logged in",
				},
				queueMessage,
			),
		)()
	}
}
