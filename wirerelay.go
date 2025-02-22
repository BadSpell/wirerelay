package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"os"
)

func generateRegistrationToken() ([]byte, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func runServer(listenAddr string, tokenB64 string) {
	var regToken []byte
	var err error
	if tokenB64 == "" {
		regToken, err = generateRegistrationToken()
		if err != nil {
			fmt.Println("Error generating registration token:", err)
			os.Exit(1)
		}
		tokenB64 = base64.StdEncoding.EncodeToString(regToken)
	} else {
		regToken, err = base64.StdEncoding.DecodeString(tokenB64)
		if err != nil {
			fmt.Println("Error decoding registration token:", err)
			os.Exit(1)
		}
		if len(regToken) != 32 {
			fmt.Println("Registration token must be 32 bytes after decoding.")
			os.Exit(1)
		}
	}
	fmt.Println("Registration token (base64):", tokenB64)

	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Relay server listening on", listenAddr)

	var relayClientAddr *net.UDPAddr
	var clientAddr *net.UDPAddr
	buf := make([]byte, 2048)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}
		data := buf[:n]

		if len(data) == 32 && bytes.Equal(data, regToken) {
			relayClientAddr = remoteAddr
			fmt.Println("Registered relay client:", remoteAddr.String())

			_, err = conn.WriteToUDP(data, remoteAddr)
		}

		if relayClientAddr != nil && remoteAddr.String() == relayClientAddr.String() {
			if clientAddr != nil {
				_, err = conn.WriteToUDP(data, clientAddr)
				if err != nil {
					fmt.Println("Error sending data to client:", err)
				}
			}
		} else {
			clientAddr = remoteAddr
			if relayClientAddr != nil {
				_, err = conn.WriteToUDP(data, relayClientAddr)
				if err != nil {
					fmt.Println("Error sending data to relay client:", err)
				}
			}
		}
	}
}

func runClient(serverAddrStr, targetAddrStr, tokenB64 string) {
	regToken, err := base64.StdEncoding.DecodeString(tokenB64)
	if err != nil {
		fmt.Println("Error decoding registration token:", err)
		os.Exit(1)
	}
	if len(regToken) != 32 {
		fmt.Println("Registration token must be 32 bytes after decoding.")
		os.Exit(1)
	}

	serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr)
	if err != nil {
		fmt.Println("Error resolving relay server address:", err)
		os.Exit(1)
	}

	targetAddr, err := net.ResolveUDPAddr("udp", targetAddrStr)
	if err != nil {
		fmt.Println("Error resolving target address:", err)
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error connecting to relay server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	_, err = conn.Write(regToken)
	if err != nil {
		fmt.Println("Error sending registration token:", err)
		os.Exit(1)
	}

	buf := make([]byte, 2048)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("Error reading from relay server:", err)
		os.Exit(1)
	}
	if !bytes.Equal(buf[:n], regToken) {
		fmt.Println("Received token does not match the registration token.")
		os.Exit(1)
	}
	fmt.Println("Registered success with relay server.")

	targetConn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		fmt.Println("Error connecting to target server:", err)
		os.Exit(1)
	}
	defer targetConn.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println("Error reading from relay server:", err)
				continue
			}
			data := buf[:n]
			_, err = targetConn.Write(data)
			if err != nil {
				fmt.Println("Error forwarding data to target server:", err)
			}
		}
	}()

	for {
		n, err := targetConn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from target server:", err)
			continue
		}
		data := buf[:n]
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Error sending data to relay server:", err)
		}
	}
}

func main() {
	serverFlag := flag.String(
		"server",
		"",
		"Run in server mode: bind address (e.g., 0.0.0.0:12345)",
	)
	clientFlag := flag.String(
		"client",
		"",
		"Run in client mode: relay server address (e.g., public.server:12345)",
	)
	targetFlag := flag.String(
		"target",
		"",
		"Target server address (e.g., 127.0.0.1:51820)",
	)
	tokenFlag := flag.String("token", "", "Registration token in base64")
	flag.Parse()

	if *serverFlag == "" && *clientFlag == "" {
		fmt.Println("Usage: you must specify either --server or --client flag.")
		flag.Usage()
		os.Exit(1)
	}

	if *serverFlag != "" && *clientFlag != "" {
		fmt.Println("Cannot run in both server and client mode simultaneously.")
		flag.Usage()
		os.Exit(1)
	}

	if *serverFlag != "" {
		runServer(*serverFlag, *tokenFlag)
	} else {
		if *tokenFlag == "" {
			fmt.Println("Client mode requires --token flag with the registration token.")
			flag.Usage()
			os.Exit(1)
		}
		runClient(*clientFlag, *targetFlag, *tokenFlag)
	}
}
