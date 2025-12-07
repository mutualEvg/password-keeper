// Package main is the entry point for the GophKeeper CLI client.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/ar11/gophkeeper/internal/client"
	"github.com/ar11/gophkeeper/internal/models"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// Version is the application version (set at build time).
	Version = "dev"
	// BuildDate is the build date (set at build time).
	BuildDate = "unknown"

	// Global flags
	serverAddress string
	gophClient    *client.Client
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper - Secure Password Manager",
		Long:  `GophKeeper is a secure client-server password manager for storing credentials, text, binary data, and bank cards.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip connection for version command
			if cmd.Name() == "version" {
				return
			}

			var err error
			gophClient, err = client.NewClient(serverAddress)
			if err != nil {
				log.Fatalf("Failed to create client: %v", err)
			}

			if err := gophClient.Connect(); err != nil {
				log.Fatalf("Failed to connect to server: %v", err)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if gophClient != nil {
				gophClient.Close()
			}
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&serverAddress, "server", "s", "localhost:50051", "Server address")

	// Add commands
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(registerCmd())
	rootCmd.AddCommand(loginCmd())
	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(getCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(syncCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// versionCmd returns the version command.
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GophKeeper CLI Client\n")
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Build Date: %s\n", BuildDate)
		},
	}
}

// registerCmd returns the register command.
func registerCmd() *cobra.Command {
	var username, password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		Run: func(cmd *cobra.Command, args []string) {
			// Prompt for password if not provided
			if password == "" {
				fmt.Print("Enter password: ")
				passBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					log.Fatalf("Failed to read password: %v", err)
				}
				password = string(passBytes)
				fmt.Println()
			}

			if err := gophClient.Register(username, password); err != nil {
				log.Fatalf("Registration failed: %v", err)
			}

			fmt.Println("Registration successful!")
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Username (email)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	cmd.MarkFlagRequired("username")

	return cmd
}

// loginCmd returns the login command.
func loginCmd() *cobra.Command {
	var username, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to the server",
		Run: func(cmd *cobra.Command, args []string) {
			// Prompt for password if not provided
			if password == "" {
				fmt.Print("Enter password: ")
				passBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					log.Fatalf("Failed to read password: %v", err)
				}
				password = string(passBytes)
				fmt.Println()
			}

			if err := gophClient.Login(username, password); err != nil {
				log.Fatalf("Login failed: %v", err)
			}

			gophClient.SetMasterPassword(password)
			fmt.Println("Login successful!")
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Username (email)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	cmd.MarkFlagRequired("username")

	return cmd
}

// addCmd returns the add command with subcommands.
func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new item",
	}

	cmd.AddCommand(addCredentialCmd())
	cmd.AddCommand(addTextCmd())
	cmd.AddCommand(addBinaryCmd())
	cmd.AddCommand(addCardCmd())

	return cmd
}

// addCredentialCmd returns the add credential subcommand.
func addCredentialCmd() *cobra.Command {
	var name, login, password, metadata string

	cmd := &cobra.Command{
		Use:   "credential",
		Short: "Add a new credential (login/password)",
		Run: func(cmd *cobra.Command, args []string) {
			// Prompt for master password
			masterPass := promptMasterPassword()
			gophClient.SetMasterPassword(masterPass)

			// Parse metadata
			meta := parseMetadata(metadata)

			if err := gophClient.AddCredential(name, login, password, meta); err != nil {
				log.Fatalf("Failed to add credential: %v", err)
			}

			fmt.Println("Credential added successfully!")
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Item name")
	cmd.Flags().StringVarP(&login, "login", "l", "", "Login")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "", "Metadata (key=value,key2=value2)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("login")
	cmd.MarkFlagRequired("password")

	return cmd
}

// addTextCmd returns the add text subcommand.
func addTextCmd() *cobra.Command {
	var name, data, metadata string

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Add text data",
		Run: func(cmd *cobra.Command, args []string) {
			// Prompt for master password
			masterPass := promptMasterPassword()
			gophClient.SetMasterPassword(masterPass)

			// Parse metadata
			meta := parseMetadata(metadata)

			if err := gophClient.AddText(name, data, meta); err != nil {
				log.Fatalf("Failed to add text: %v", err)
			}

			fmt.Println("Text data added successfully!")
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Item name")
	cmd.Flags().StringVarP(&data, "data", "d", "", "Text content")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "", "Metadata (key=value,key2=value2)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("data")

	return cmd
}

// addBinaryCmd returns the add binary subcommand.
func addBinaryCmd() *cobra.Command {
	var name, file, metadata string

	cmd := &cobra.Command{
		Use:   "binary",
		Short: "Add binary data from file",
		Run: func(cmd *cobra.Command, args []string) {
			// Prompt for master password
			masterPass := promptMasterPassword()
			gophClient.SetMasterPassword(masterPass)

			// Read file
			data, err := os.ReadFile(file)
			if err != nil {
				log.Fatalf("Failed to read file: %v", err)
			}

			// Parse metadata
			meta := parseMetadata(metadata)

			if err := gophClient.AddBinary(name, file, data, meta); err != nil {
				log.Fatalf("Failed to add binary data: %v", err)
			}

			fmt.Println("Binary data added successfully!")
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Item name")
	cmd.Flags().StringVarP(&file, "file", "f", "", "File path")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "", "Metadata (key=value,key2=value2)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("file")

	return cmd
}

// addCardCmd returns the add card subcommand.
func addCardCmd() *cobra.Command {
	var name, number, holder, cvv, expiry, metadata string

	cmd := &cobra.Command{
		Use:   "card",
		Short: "Add bank card data",
		Run: func(cmd *cobra.Command, args []string) {
			// Prompt for master password
			masterPass := promptMasterPassword()
			gophClient.SetMasterPassword(masterPass)

			// Parse expiry (MM/YY)
			parts := strings.Split(expiry, "/")
			if len(parts) != 2 {
				log.Fatal("Expiry must be in format MM/YY")
			}

			card := models.CardData{
				Number:      number,
				Holder:      holder,
				CVV:         cvv,
				ExpiryMonth: parts[0],
				ExpiryYear:  parts[1],
			}

			// Parse metadata
			meta := parseMetadata(metadata)

			if err := gophClient.AddCard(name, card, meta); err != nil {
				log.Fatalf("Failed to add card: %v", err)
			}

			fmt.Println("Card data added successfully!")
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Item name")
	cmd.Flags().StringVar(&number, "number", "", "Card number")
	cmd.Flags().StringVar(&holder, "holder", "", "Cardholder name")
	cmd.Flags().StringVar(&cvv, "cvv", "", "CVV")
	cmd.Flags().StringVar(&expiry, "expiry", "", "Expiry date (MM/YY)")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "", "Metadata (key=value,key2=value2)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("number")

	return cmd
}

// getCmd returns the get command.
func getCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an item by name",
		Run: func(cmd *cobra.Command, args []string) {
			// Prompt for master password
			masterPass := promptMasterPassword()
			gophClient.SetMasterPassword(masterPass)

			item, data, err := gophClient.GetItem(name)
			if err != nil {
				log.Fatalf("Failed to get item: %v", err)
			}

			fmt.Printf("Name: %s\n", item.Name)
			fmt.Printf("Type: %s\n", item.Type)
			fmt.Printf("Created: %s\n", item.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated: %s\n", item.UpdatedAt.Format("2006-01-02 15:04:05"))

			if len(item.Metadata) > 0 {
				fmt.Println("Metadata:")
				for k, v := range item.Metadata {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}

			fmt.Println("\nData:")
			switch d := data.(type) {
			case models.Credential:
				fmt.Printf("  Login: %s\n", d.Login)
				fmt.Printf("  Password: %s\n", d.Password)
			case models.TextData:
				fmt.Printf("  Content: %s\n", d.Content)
			case models.BinaryData:
				fmt.Printf("  Filename: %s\n", d.Filename)
				fmt.Printf("  Size: %d bytes\n", len(d.Data))
			case models.CardData:
				fmt.Printf("  Number: %s\n", maskCard(d.Number))
				fmt.Printf("  Holder: %s\n", d.Holder)
				fmt.Printf("  CVV: %s\n", d.CVV)
				fmt.Printf("  Expiry: %s/%s\n", d.ExpiryMonth, d.ExpiryYear)
			}
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Item name")
	cmd.MarkFlagRequired("name")

	return cmd
}

// listCmd returns the list command.
func listCmd() *cobra.Command {
	var dataType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all items",
		Run: func(cmd *cobra.Command, args []string) {
			// Prompt for master password
			masterPass := promptMasterPassword()
			gophClient.SetMasterPassword(masterPass)

			var dt models.DataType
			if dataType != "" {
				dt = models.DataType(dataType)
			}

			items, err := gophClient.ListItems(dt)
			if err != nil {
				log.Fatalf("Failed to list items: %v", err)
			}

			if len(items) == 0 {
				fmt.Println("No items found")
				return
			}

			fmt.Printf("Found %d item(s):\n\n", len(items))
			for _, item := range items {
				fmt.Printf("Name: %s\n", item.Name)
				fmt.Printf("Type: %s\n", item.Type)
				fmt.Printf("Created: %s\n", item.CreatedAt.Format("2006-01-02 15:04:05"))
				if len(item.Metadata) > 0 {
					fmt.Printf("Metadata: ")
					first := true
					for k, v := range item.Metadata {
						if !first {
							fmt.Printf(", ")
						}
						fmt.Printf("%s=%s", k, v)
						first = false
					}
					fmt.Println()
				}
				fmt.Println("---")
			}
		},
	}

	cmd.Flags().StringVarP(&dataType, "type", "t", "", "Filter by type (credential, text, binary, card)")

	return cmd
}

// syncCmd returns the sync command.
func syncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data with server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := gophClient.Sync(); err != nil {
				log.Fatalf("Sync failed: %v", err)
			}

			fmt.Println("Synchronization successful!")
		},
	}
}

// Helper functions

func promptMasterPassword() string {
	fmt.Print("Enter master password: ")
	passBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	fmt.Println()
	return string(passBytes)
}

func parseMetadata(metadata string) map[string]string {
	result := make(map[string]string)
	if metadata == "" {
		return result
	}

	pairs := strings.Split(metadata, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return result
}

func maskCard(number string) string {
	if len(number) <= 4 {
		return number
	}
	return "****" + number[len(number)-4:]
}

