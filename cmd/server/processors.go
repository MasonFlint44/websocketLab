package main

// TODO: update documentation
// TODO: test updated processors

func sendMessages() {
	for {
		message := <-outboundResponses
		clientPipe(message.GetClient(), nil,
			hasClient,
			hasConn,
			onClientError(
				sendMessageToClient(message),
				toClientErrorHandler(printError),
			),
		)
	}
}

func processLogoutRequests() {
	for {
		req := <-logoutRequests
		clientPipe(req.GetClient(), nil,
			hasClient,
			hasConn,
			onClientError(
				hasAuth,
				clientProcessorToErrorHandler(queueCustomMessageToClient("Server", "Client is not logged in")),
			),
			logout,
			queueCustomMessageToClient("Server", "Successful logout"),
		)
	}
}

func processNewUserRequests() {
	for {
		req := <-newUserRequests
		client, err := clientPipe(req.GetClient(), nil,
			hasClient,
			hasConn,
		)
		message, err := messagePipe(req.GetMessage(), err,
			hasMessage,
		)
		client, err = clientPipe(message.GetClient(), err,
			hasClient,
			setConn(client),
			onClientError(
				validHandle,
				clientProcessorToErrorHandler(queueCustomMessageToClient("Server", "Handle must be less than 32 characters")),
			),
			onClientError(
				validPass,
				clientProcessorToErrorHandler(queueCustomMessageToClient("Server", "Pass must be between 4 and 8 characters")),
			),
			onClientError(
				uniqueHandle,
				clientProcessorToErrorHandler(queueCustomMessageToClient("Server", "Handle is already taken")),
			),
			onClientError(
				register,
				clientProcessorToErrorHandler(queueCustomMessageToClient("Server", "Unable to register new user")),
			),
			queueCustomMessageToClient("Server", "Welcome! Use 'login' to continue."),
		)
	}
}

func processSendRequests() {
	for {
		req := <-sendRequests
		client, err := clientPipe(req.GetClient(), nil,
			hasClient,
			hasConn,
			onClientError(
				hasAuth,
				clientProcessorToErrorHandler(queueCustomMessageToClient("Server", "Unauthorized - Please login")),
			),
		)
		err = forEachClient(err,
			setHandle(client),
			queueMessageToClient(req.GetMessage()),
		)
	}
}

func processLoginRequests() {
	for {
		req := <-loginRequests
		message, err := messagePipe(req.GetMessage(), nil,
			hasMessage,
		)
		messageClient, err := clientPipe(message.GetClient(), err,
			hasClient,
		)
		_, err = clientPipe(req.GetClient(), nil,
			hasClient,
			hasConn,
			onClientError(
				hasAuth,
				clientProcessorToErrorHandler(onClientError(
					authorize(messageClient),
					clientProcessorToErrorHandler(queueCustomMessageToClient("Server", "Unable to log in with provided credentials")),
				)),
				clientProcessorToErrorHandler(queueCustomMessageToClient("Server", "Successful login")),
			),
			queueCustomMessageToClient("Server", "Client is already logged in"),
		)
	}
}
