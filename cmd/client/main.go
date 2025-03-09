package main

import (
	"flag"
	"fmt"
	"log"
	"pcbook/client"
	"pcbook/pb"
	"pcbook/sample"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const refreshDuration = 30 * time.Second

func authMethods() map[string]bool {
	return map[string]bool{
		"/pcbook.pcbook.LaptopService/CreateLaptop": true,
		"/pcbook.pcbook.LaptopService/UploadImage":  true,
		"/pcbook.pcbook.LaptopService/RateLaptop":   true,
	}
}

func main() {
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dial server on address %s", *serverAddress)

	cc1, err := grpc.NewClient(
		*serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("cannot dial server: %v", err)
	}
	defer conn.Close()

	log.Print("create laptop client")
	authClient := client.NewAuthClient(cc1, "admin", "secret")
	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration) := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)
	if err != nil {
		log.Fatal("cannot create auth interceptor: ", err)
	}

	cc2, err := grpc.NewClient(
		*serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatalf("cannot dial server: %v", err)
	}
	defer conn.Close()
	
	laptopClient := client.NewLaptopClient(cc2)
	testRateLaptop(laptopClient)
}

func testCreateLaptop(laptopClient *client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(laptop)
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	laptopClient.SearchLaptop(filter)
}

func testUploadImage(laptopClient *client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.Id, "tmp/laptop.jpg")
}

func testRateLaptop(laptopClient *client.LaptopClient) {
	n := 3
	laptopIDs := make([]string, n)
	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopClient.CreateLaptop(laptop)
		laptopIDs[i] = laptop.Id
	}
	scores := make([]float64, n)
	for {
		fmt.Println("rate laptop (y/n)?")
		var answer string
		fmt.Scan(&answer)
		if strings.ToLower(answer) == "n" {
			break
		}
		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}
		err := laptopClient.RateLaptop(laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}
