# Loadtesting using Golang self-deploying apps

## Setup

-   Make sure you have go installed
-   On mac you can do this with `$ brew install go`

### Server

-   Build and run the server locally with `scripts/run-server-local`
-   Check that its working by going to `localhost` in your browser
-   You should see the current time and date

### Loadtest

-   Get cluster credentials and change context with `gcloud container clusters get-credentials women-who-go-demo`
-   Run the loadtest locally with `$ scripts/run-loadtest`
-   Run the loadtest on kubernetes with `$ scripts/run-loadtest --kubernetes --replicas=<num-replicas>`
