package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/Rbm3267/wingllm/internal/discovery"
	"github.com/Rbm3267/wingllm/internal/engine"
	"github.com/Rbm3267/wingllm/internal/proxy"
	"github.com/Rbm3267/wingllm/internal/vram"
)

var banner = `
 __      __.__             .____    .____       _____   
/  \    /  \__| ____    ____|    |   |    |     /     \  
\   \/\/   /  |/    \  / ___/    |   |    |    /  \ /  \ 
 \        /|  |   |  \/ /_/  >    |___|    |___/    Y    \
  \__/\  / |__|___|  /\___  /|_______ \_______ \____|__  /
       \/          \//_____/         \/       \/       \/ 
   Heterogeneous Speculative Decoding for Local AI
`

func main() {
	if len(os.Args) < 2 {
		handleInteractive()
		return
	}

	command := os.Args[1]

	switch command {
	case "vram":
		handleVRAM()
	case "list":
		handleList()
	case "serve", "run":
		handleServe(os.Args[2:])
	default:
		fmt.Printf("Unknown command '%s'\n", command)
		printUsage()
	}
}

func handleInteractive() {
	fmt.Println(banner)
	fmt.Println(" Scanning local models in Ollama...")

	models, err := discovery.ScanOllamaModels()
	if err != nil || len(models) == 0 {
		fmt.Println("No local Ollama models found.")
		fmt.Print("Would you like to auto-pull Qwen 2.5 Coder 7B + 0.5B draft? (Y/n): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" || input == "y" || input == "yes" {
			pullCmd := exec.Command("ollama", "pull", "qwen2.5-coder:7b")
			pullCmd.Stdout = os.Stdout
			pullCmd.Stderr = os.Stderr
			pullCmd.Run()
			models, _ = discovery.ScanOllamaModels()
		} else {
			return
		}
	}

	// Filter viable target models
	var viableTargets []string
	for _, m := range models {
		if !m.IsDraft {
			viableTargets = append(viableTargets, m.Name)
		}
	}

	if len(viableTargets) == 0 {
		fmt.Println("No target models available. Run 'ollama pull qwen2.5-coder:7b'")
		return
	}

	fmt.Println("\nAvailable models ready for 2x Speculative Turbo:")
	for i, name := range viableTargets {
		rec := ""
		if i == 0 || strings.Contains(name, "7b") {
			rec = " [Recommended]"
		}
		fmt.Printf("  [%d] %-30s%s\n", i+1, name, rec)
	}

	fmt.Print("\nSelect model number (or press Enter for [1]): ")
	reader := bufio.NewReader(os.Stdin)
	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)

	choiceIndex := 0
	if choiceStr != "" {
		if num, err := strconv.Atoi(choiceStr); err == nil && num >= 1 && num <= len(viableTargets) {
			choiceIndex = num - 1
		}
	}

	selectedModel := viableTargets[choiceIndex]
	fmt.Printf("\nLaunching WingLLM with '%s'...\n\n", selectedModel)
	handleServe([]string{selectedModel})
}

func printUsage() {
	fmt.Println(banner)
	fmt.Println("Usage: wingllm <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  serve <model>   Start speculative inference server (auto-pairs with draft model)")
	fmt.Println("  list            Scan local Ollama models and show available target/draft pairs")
	fmt.Println("  vram            Inspect GPU VRAM status and available budget")
	fmt.Println("\nExamples:")
	fmt.Println("  wingllm serve qwen2.5-coder:7b")
	fmt.Println("  wingllm list")
	fmt.Println("  wingllm vram")
}

func handleVRAM() {
	status, err := vram.GetVRAMStatus()
	if err != nil {
		fmt.Printf("Error querying VRAM: %v\n", err)
		return
	}

	fmt.Println("\n--- Local GPU VRAM Telemetry ---")
	fmt.Printf("GPU Device:   %s\n", status.GPUName)
	fmt.Printf("Total VRAM:   %d MB\n", status.TotalMB)
	fmt.Printf("Used VRAM:    %d MB\n", status.UsedMB)
	fmt.Printf("Free VRAM:    %d MB\n", status.FreeMB)
	if status.SafeFor2GB {
		fmt.Println("Safety Check:  Sufficient VRAM headroom for 0.5B draft model")
	} else {
		fmt.Println("Safety Check: ⚠ VRAM headroom tight (<400MB free). Consider closing browser tabs.")
	}
	fmt.Println()
}

func handleList() {
	models, err := discovery.ScanOllamaModels()
	if err != nil {
		fmt.Printf("Error scanning Ollama models: %v\n", err)
		return
	}

	fmt.Println("\n--- Discovered Local Models ---")
	var targets []discovery.ModelInfo
	var drafts []discovery.ModelInfo

	for _, m := range models {
		if m.IsDraft {
			drafts = append(drafts, m)
		} else {
			targets = append(targets, m)
		}
	}

	fmt.Println("\n[TARGET MODELS - CPU System RAM]:")
	for _, t := range targets {
		fmt.Printf("  • %-30s (%.1f GB) [Family: %s]\n", t.Name, float64(t.Size)/(1024*1024*1024), t.Family)
	}

	fmt.Println("\n[WINGMAN DRAFT MODELS - 2GB GPU VRAM]:")
	for _, d := range drafts {
		fmt.Printf("  • %-30s (%.0f MB) [Family: %s]\n", d.Name, float64(d.Size)/(1024*1024), d.Family)
	}

	fmt.Println("\n[AUTO-RECOMMENDED PAIRINGS]:")
	for _, t := range targets {
		if pair, err := discovery.FindBestPair(t.Name, models); err == nil {
			fmt.Printf("  ✔ wingllm serve %-20s -> drafts with %s (%s)\n", t.Name, pair.Draft.Name, pair.Target.Family)
		}
	}
	fmt.Println()
}

func handleServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "Public OpenAI API port")
	enginePort := fs.Int("engine-port", 8081, "Internal llama-server port")
	ctxSize := fs.Int("ctx", 2048, "Context window size")
	draftNMax := fs.Int("draft-n", 3, "Number of tokens to draft ahead per step")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Error: Missing target model name.")
		fmt.Println("Example: wingllm serve qwen2.5-coder:7b")
		return
	}

	targetModel := fs.Arg(0)

	models, err := discovery.ScanOllamaModels()
	if err != nil {
		fmt.Printf("Error scanning local models: %v\n", err)
		return
	}

	pair, err := discovery.FindBestPair(targetModel, models)
	if err != nil {
		// Check if target model exists but needs draft model
		var target *discovery.ModelInfo
		for i := range models {
			if strings.EqualFold(models[i].Name, targetModel) || strings.Contains(strings.ToLower(models[i].Name), strings.ToLower(targetModel)) {
				target = &models[i]
				break
			}
		}
		if target != nil {
			draftName := discovery.GetDefaultDraftName(target.Family)
			fmt.Printf("\nCompatible draft model not found locally.\nAuto-pulling recommended draft model '%s' via Ollama (~397 MB)...\n\n", draftName)
			pullCmd := exec.Command("ollama", "pull", draftName)
			pullCmd.Stdout = os.Stdout
			pullCmd.Stderr = os.Stderr
			if err := pullCmd.Run(); err == nil {
				// Rescan models
				models, _ = discovery.ScanOllamaModels()
				pair, err = discovery.FindBestPair(targetModel, models)
			}
		}
		if pair == nil {
			fmt.Printf("Pairing Error: %v\n", err)
			return
		}
	}

	fmt.Println(banner)
	fmt.Printf("Target Model:  %s (Loaded into System RAM)\n", pair.Target.Name)
	fmt.Printf("Wingman Draft: %s (Offloaded to 2GB GPU)\n", pair.Draft.Name)
	fmt.Printf("Strategy:      %s\n", pair.Reason)

	cfg := engine.DefaultConfig()
	cfg.Port = *enginePort
	cfg.CtxSize = *ctxSize
	cfg.DraftNMax = *draftNMax

	sup := engine.NewSupervisor(pair, cfg)
	fmt.Println("\nStarting engine supervisor...")
	if err := sup.Start(); err != nil {
		fmt.Printf("Failed to start engine: %v\n", err)
		return
	}
	defer sup.Stop()

	srv, err := proxy.NewServer(*port, *enginePort, pair)
	if err != nil {
		fmt.Printf("Failed to create proxy server: %v\n", err)
		return
	}

	// Graceful shutdown on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down WingLLM cleanly...")
		sup.Stop()
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
