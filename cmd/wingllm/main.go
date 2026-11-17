package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bennett/wingllm/internal/discovery"
	"github.com/Bennett/wingllm/internal/engine"
	"github.com/Bennett/wingllm/internal/proxy"
	"github.com/Bennett/wingllm/internal/vram"
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
		printUsage()
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
		fmt.Printf("Pairing Error: %v\n", err)
		return
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
