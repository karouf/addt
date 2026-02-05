package cmd

import (
	"fmt"
	"os"
	"strings"

	extcmd "github.com/jedi4ever/addt/cmd/extensions"
)

// HandleCompletionCommand generates shell completion scripts
func HandleCompletionCommand(args []string) {
	if len(args) == 0 {
		printCompletionHelp()
		return
	}

	shell := args[0]
	switch shell {
	case "bash":
		fmt.Print(bashCompletion())
	case "zsh":
		fmt.Print(zshCompletion())
	case "fish":
		fmt.Print(fishCompletion())
	case "-h", "--help", "help":
		printCompletionHelp()
	default:
		fmt.Printf("Unknown shell: %s\n", shell)
		fmt.Println("Supported shells: bash, zsh, fish")
		os.Exit(1)
	}
}

func printCompletionHelp() {
	fmt.Println("Generate shell completion scripts")
	fmt.Println()
	fmt.Println("Usage: addt completion <shell>")
	fmt.Println()
	fmt.Println("Supported shells:")
	fmt.Println("  bash    Generate bash completion script")
	fmt.Println("  zsh     Generate zsh completion script")
	fmt.Println("  fish    Generate fish completion script")
	fmt.Println()
	fmt.Println("Setup:")
	fmt.Println()
	fmt.Println("  Bash (add to ~/.bashrc):")
	fmt.Println("    eval \"$(addt completion bash)\"")
	fmt.Println()
	fmt.Println("  Zsh (add to ~/.zshrc):")
	fmt.Println("    eval \"$(addt completion zsh)\"")
	fmt.Println()
	fmt.Println("  Fish (run once):")
	fmt.Println("    addt completion fish > ~/.config/fish/completions/addt.fish")
}

// getExtensionNames returns available extension names for completion
func getExtensionNames() []string {
	extensions := extcmd.ListExtensions()
	var names []string
	for _, ext := range extensions {
		names = append(names, ext.Name)
	}
	return names
}

func bashCompletion() string {
	extensions := strings.Join(getExtensionNames(), " ")

	return fmt.Sprintf(`# addt bash completion
_addt_completions() {
    local cur prev words cword
    _init_completion || return

    local commands="run build shell containers config extensions firewall completion doctor version cli"
    local config_cmds="global project extension"
    local containers_cmds="list clean"
    local firewall_cmds="global project"
    local firewall_actions="list allow deny remove"
    local extensions_cmds="list info new"
    local extensions="%s"
    local config_keys="persistent ports ssh_forward gpg_forward dind firewall firewall_mode docker_cpus docker_memory workdir workdir_automount workdir_readonly provider node_version go_version uv_version"

    case "${cword}" in
        1)
            COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
            ;;
        2)
            case "${prev}" in
                run|build|shell)
                    COMPREPLY=($(compgen -W "${extensions}" -- "${cur}"))
                    ;;
                config)
                    COMPREPLY=($(compgen -W "${config_cmds}" -- "${cur}"))
                    ;;
                containers)
                    COMPREPLY=($(compgen -W "${containers_cmds}" -- "${cur}"))
                    ;;
                firewall)
                    COMPREPLY=($(compgen -W "${firewall_cmds}" -- "${cur}"))
                    ;;
                extensions)
                    COMPREPLY=($(compgen -W "${extensions_cmds}" -- "${cur}"))
                    ;;
                completion)
                    COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
                    ;;
            esac
            ;;
        3)
            case "${words[1]}" in
                config)
                    case "${prev}" in
                        global|project)
                            COMPREPLY=($(compgen -W "list set unset" -- "${cur}"))
                            ;;
                        extension)
                            COMPREPLY=($(compgen -W "${extensions}" -- "${cur}"))
                            ;;
                    esac
                    ;;
                firewall)
                    COMPREPLY=($(compgen -W "${firewall_actions}" -- "${cur}"))
                    ;;
                extensions)
                    case "${prev}" in
                        info)
                            COMPREPLY=($(compgen -W "${extensions}" -- "${cur}"))
                            ;;
                    esac
                    ;;
            esac
            ;;
        4)
            case "${words[1]}" in
                config)
                    case "${words[3]}" in
                        set|unset)
                            COMPREPLY=($(compgen -W "${config_keys}" -- "${cur}"))
                            ;;
                    esac
                    ;;
            esac
            ;;
    esac
}

complete -F _addt_completions addt
`, extensions)
}

func zshCompletion() string {
	extensions := strings.Join(getExtensionNames(), " ")

	return fmt.Sprintf(`#compdef addt

_addt() {
    local -a commands extensions config_cmds containers_cmds firewall_cmds firewall_actions extensions_cmds config_keys

    commands=(
        'run:Run an agent in a container'
        'build:Build container image for an agent'
        'shell:Open a shell in a container'
        'containers:Manage containers'
        'config:Manage configuration'
        'extensions:Manage extensions'
        'firewall:Manage firewall rules'
        'completion:Generate shell completions'
        'doctor:Check system health'
        'version:Show version information'
        'cli:CLI management commands'
    )

    extensions=(%s)

    config_cmds=(
        'global:Manage global configuration'
        'project:Manage project configuration'
        'extension:Manage extension configuration'
    )

    containers_cmds=(
        'list:List containers'
        'clean:Remove all addt containers'
    )

    firewall_cmds=(
        'global:Manage global firewall rules'
        'project:Manage project firewall rules'
    )

    firewall_actions=(
        'list:List firewall rules'
        'allow:Allow a domain'
        'deny:Deny a domain'
        'remove:Remove a rule'
    )

    extensions_cmds=(
        'list:List available extensions'
        'info:Show extension details'
        'new:Create a new extension'
    )

    config_keys=(
        'persistent' 'ports' 'ssh_forward' 'gpg_forward' 'dind'
        'firewall' 'firewall_mode' 'docker_cpus' 'docker_memory'
        'workdir' 'workdir_automount' 'workdir_readonly' 'provider'
        'node_version' 'go_version' 'uv_version'
    )

    _arguments -C \
        '1: :->command' \
        '2: :->subcommand' \
        '3: :->arg3' \
        '4: :->arg4' \
        '*::arg:->args'

    case "$state" in
        command)
            _describe -t commands 'addt commands' commands
            ;;
        subcommand)
            case "$words[2]" in
                run|build|shell)
                    _describe -t extensions 'extensions' extensions
                    ;;
                config)
                    _describe -t config_cmds 'config commands' config_cmds
                    ;;
                containers)
                    _describe -t containers_cmds 'container commands' containers_cmds
                    ;;
                firewall)
                    _describe -t firewall_cmds 'firewall commands' firewall_cmds
                    ;;
                extensions)
                    _describe -t extensions_cmds 'extension commands' extensions_cmds
                    ;;
                completion)
                    _values 'shell' 'bash' 'zsh' 'fish'
                    ;;
            esac
            ;;
        arg3)
            case "$words[2]" in
                config)
                    case "$words[3]" in
                        global|project)
                            _values 'action' 'list' 'set' 'unset'
                            ;;
                        extension)
                            _describe -t extensions 'extensions' extensions
                            ;;
                    esac
                    ;;
                firewall)
                    _describe -t firewall_actions 'firewall actions' firewall_actions
                    ;;
                extensions)
                    case "$words[3]" in
                        info)
                            _describe -t extensions 'extensions' extensions
                            ;;
                    esac
                    ;;
            esac
            ;;
        arg4)
            case "$words[2]" in
                config)
                    case "$words[4]" in
                        set|unset)
                            _describe -t config_keys 'config keys' config_keys
                            ;;
                    esac
                    ;;
            esac
            ;;
    esac
}

_addt "$@"
`, extensions)
}

func fishCompletion() string {
	extensions := getExtensionNames()

	var sb strings.Builder
	sb.WriteString("# addt fish completion\n\n")

	// Disable file completion by default
	sb.WriteString("complete -c addt -f\n\n")

	// Main commands
	sb.WriteString("# Main commands\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'run' -d 'Run an agent in a container'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'build' -d 'Build container image for an agent'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'shell' -d 'Open a shell in a container'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'containers' -d 'Manage containers'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'config' -d 'Manage configuration'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'extensions' -d 'Manage extensions'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'firewall' -d 'Manage firewall rules'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'completion' -d 'Generate shell completions'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'doctor' -d 'Check system health'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'version' -d 'Show version information'\n")
	sb.WriteString("complete -c addt -n '__fish_use_subcommand' -a 'cli' -d 'CLI management commands'\n")
	sb.WriteString("\n")

	// Extensions for run/build/shell
	sb.WriteString("# Extensions\n")
	for _, ext := range extensions {
		sb.WriteString(fmt.Sprintf("complete -c addt -n '__fish_seen_subcommand_from run build shell' -a '%s'\n", ext))
	}
	sb.WriteString("\n")

	// Config subcommands
	sb.WriteString("# Config subcommands\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from config' -a 'global' -d 'Manage global configuration'\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from config' -a 'project' -d 'Manage project configuration'\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from config' -a 'extension' -d 'Manage extension configuration'\n")
	sb.WriteString("\n")

	// Containers subcommands
	sb.WriteString("# Containers subcommands\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from containers' -a 'list' -d 'List containers'\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from containers' -a 'clean' -d 'Remove all addt containers'\n")
	sb.WriteString("\n")

	// Firewall subcommands
	sb.WriteString("# Firewall subcommands\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from firewall' -a 'global' -d 'Manage global firewall rules'\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from firewall' -a 'project' -d 'Manage project firewall rules'\n")
	sb.WriteString("\n")

	// Extensions subcommands
	sb.WriteString("# Extensions subcommands\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from extensions' -a 'list' -d 'List available extensions'\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from extensions' -a 'info' -d 'Show extension details'\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from extensions' -a 'new' -d 'Create a new extension'\n")
	sb.WriteString("\n")

	// Completion subcommands
	sb.WriteString("# Completion shells\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from completion' -a 'bash' -d 'Generate bash completion'\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from completion' -a 'zsh' -d 'Generate zsh completion'\n")
	sb.WriteString("complete -c addt -n '__fish_seen_subcommand_from completion' -a 'fish' -d 'Generate fish completion'\n")

	return sb.String()
}
