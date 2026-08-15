package tools

// registerAll registers every built-in tool and its generator.
func (c *Catalog) registerAll() {
	c.register(nmapTool, generateNmap)
	c.register(gobusterTool, generateGobuster)
	c.register(niktoTool, generateNikto)
	c.register(sqlmapTool, generateSQLMap)
	c.register(metasploitTool, generateMetasploit)
	c.register(hydraTool, generateHydra)
	c.register(hashcatTool, generateHashcat)
	c.register(aircrackTool, generateAircrack)
	c.register(tcpdumpTool, generateTCPDump)
	c.register(impacketTool, generateImpacket)
	c.register(mimikatzTool, generateMimikatz)
	c.register(caidoTool, generateCaido)
	c.register(bloodhoundTool, generateBloodHound)
	c.register(fluxionTool, generateFluxion)
	c.register(ffufTool, generateFFUF)
	c.register(msfvenomTool, generateMsfvenom)
	c.register(masscanTool, generateMasscan)
	c.register(beefTool, generateBeEF)
	c.register(kismetTool, generateKismet)
	c.register(xsserTool, generateXsser)
	c.register(commixTool, generateCommix)
	c.register(sherlockTool, generateSherlock)
	c.register(holeheTool, generateHolehe)
	c.register(rdnsTool, generateRDNS)
}

func (c *Catalog) register(t *Tool, g Generator) {
	c.tools = append(c.tools, t)
	c.generators[t.ID] = g
}

// Category ids used across the catalog.
const (
	CatRecon       = "recon"
	CatWeb         = "web"
	CatExploit     = "exploitation"
	CatCreds       = "credentials"
	CatWireless    = "wireless"
	CatNetwork     = "network"
	CatPostExploit = "postexploit"
)

var nmapTool = &Tool{
	ID:          "nmap",
	Name:        "Nmap",
	Icon:        "🗺️",
	Description: "O scanner de rede mais usado no mundo. Descoberta de hosts, portas, serviços, sistema operacional e scripts de vulnerabilidade.",
	Category:    CatRecon,
	Risk:        "medium",
	Install:     "sudo apt install nmap",
	Docs:        "https://nmap.org/docs.html",
	Tags:        []string{"rede", "scanner", "portas", "enumeração"},
	Questions: []Question{
		{
			ID:          "targets",
			Label:       "Alvos (IP, host ou rede)",
			Type:        "text",
			Placeholder: "10.10.10.0/24 ou 192.168.1.10",
			Help:        "Pode ser um único IP, um range (10.0.0.1-50) ou uma rede CIDR (10.10.10.0/24).",
			Required:    true,
		},
		{
			ID:      "scanType",
			Label:   "Tipo de varredura",
			Type:    "select",
			Default: "syn",
			Options: []Option{
				{Value: "syn", Label: "SYN (-sS) — rápido e silencioso"},
				{Value: "connect", Label: "Connect (-sT) — usa TCP completo"},
				{Value: "udp", Label: "UDP (-sU) — serviços UDP"},
				{Value: "version", Label: "Versão dos serviços (-sV)"},
				{Value: "aggr", Label: "Agressiva (-A) — SO + versão + scripts"},
				{Value: "vuln", Label: "Scripts de vulnerabilidade (--script vuln)"},
				{Value: "ping", Label: "Ping sweep — apenas hosts ativos"},
			},
		},
		{
			ID:          "ports",
			Label:       "Portas (opcional)",
			Type:        "text",
			Placeholder: "80,443,8080 ou 1-1000",
			Help:        "Deixe vazio para portas padrão ou use -p- para todas.",
		},
		{
			ID:    "output",
			Label: "Formatos de saída",
			Type:  "multi",
			Options: []Option{
				{Value: "normal", Label: "Texto normal (-oN)"},
				{Value: "xml", Label: "XML (-oX)"},
				{Value: "grepable", Label: "Grepable (-oG)"},
			},
		},
		{
			ID:      "verbose",
			Label:   "Modo verboso",
			Type:    "boolean",
			Default: "true",
			Help:    "Adiciona -vv com progresso detalhado.",
		},
	},
}

var gobusterTool = &Tool{
	ID:          "gobuster",
	Name:        "Gobuster",
	Icon:        "🧭",
	Description: "Enumeração de diretórios/subdomínios e vhosts via força bruta com wordlists.",
	Category:    CatWeb,
	Risk:        "low",
	Install:     "sudo apt install gobuster",
	Docs:        "https://github.com/OJ/gobuster",
	Tags:        []string{"web", "enumeração", "diretórios", "subdomínios"},
	Questions: []Question{
		{
			ID:      "mode",
			Label:   "Modo de enumeração",
			Type:    "select",
			Default: "dir",
			Options: []Option{
				{Value: "dir", Label: "Diretórios (dir)"},
				{Value: "dns", Label: "Subdomínios (dns)"},
				{Value: "vhost", Label: "Virtual hosts (vhost)"},
			},
		},
		{
			ID:          "url",
			Label:       "URL / domínio alvo",
			Type:        "text",
			Placeholder: "https://example.com ou example.com",
			Required:    true,
		},
		{
			ID:          "wordlist",
			Label:       "Wordlist",
			Type:        "text",
			Placeholder: "/usr/share/wordlists/dirb/common.txt",
			Default:     "/usr/share/wordlists/dirb/common.txt",
			Help:        "Caminho da wordlist usada na força bruta.",
		},
		{
			ID:          "extensions",
			Label:       "Extensões (apenas dir, opcional)",
			Type:        "text",
			Placeholder: "php,html,txt",
			Help:        "Separe com vírgulas. Usa -x.",
		},
		{
			ID:      "threads",
			Label:   "Número de threads",
			Type:    "number",
			Default: "40",
			Min:     1,
		},
		{
			ID:          "status",
			Label:       "Códigos de status para exibir (opcional)",
			Type:        "text",
			Placeholder: "200,301,302",
		},
	},
}

var niktoTool = &Tool{
	ID:          "nikto",
	Name:        "Nikto",
	Icon:        "🕷️",
	Description: "Scanner de servidores web que detecta arquivos perigosos, configurações erradas e softwares desatualizados.",
	Category:    CatWeb,
	Risk:        "low",
	Install:     "sudo apt install nikto",
	Docs:        "https://github.com/sullo/nikto",
	Tags:        []string{"web", "scanner", "vulnerabilidades"},
	Questions: []Question{
		{
			ID:          "url",
			Label:       "URL alvo",
			Type:        "text",
			Placeholder: "http://192.168.1.10 ou https://example.com",
			Required:    true,
		},
		{
			ID:          "port",
			Label:       "Porta (opcional)",
			Type:        "number",
			Placeholder: "8080",
			Help:        "Usa -p. Deixe vazio para a porta padrão.",
			Min:         1,
			Max:         65535,
		},
		{
			ID:    "ssl",
			Label: "Forçar SSL/TLS",
			Type:  "boolean",
			Help:  "Adiciona -ssl para testes em HTTPS.",
		},
		{
			ID:          "tuning",
			Label:       "Tuning (opcional)",
			Type:        "text",
			Placeholder: "1,2,3",
			Help:        "Categorias de teste: 1 arquivos, 2 configs erradas, 3 injeção...",
		},
		{
			ID:    "verbose",
			Label: "Modo verboso",
			Type:  "boolean",
		},
	},
}

var sqlmapTool = &Tool{
	ID:          "sqlmap",
	Name:        "SQLMap",
	Icon:        "🐍",
	Description: "Ferramenta automatizada para detecção e exploração de injeção SQL.",
	Category:    CatWeb,
	Risk:        "high",
	Install:     "sudo apt install sqlmap",
	Docs:        "https://sqlmap.org/",
	Tags:        []string{"web", "injeção", "sql", "banco de dados"},
	Questions: []Question{
		{
			ID:          "url",
			Label:       "URL da requisição",
			Type:        "text",
			Placeholder: "http://example.com/item.php?id=1",
			Required:    true,
		},
		{
			ID:      "method",
			Label:   "Método HTTP",
			Type:    "select",
			Default: "get",
			Options: []Option{
				{Value: "get", Label: "GET"},
				{Value: "post", Label: "POST"},
			},
		},
		{
			ID:          "data",
			Label:       "Corpo POST / parâmetros (opcional)",
			Type:        "text",
			Placeholder: "user=admin&pass=teste",
			Help:        "Obrigatório quando o método é POST.",
		},
		{
			ID:      "dbms",
			Label:   "Banco de dados alvo",
			Type:    "select",
			Default: "auto",
			Options: []Option{
				{Value: "auto", Label: "Auto-detecção"},
				{Value: "mysql", Label: "MySQL"},
				{Value: "postgres", Label: "PostgreSQL"},
				{Value: "mssql", Label: "Microsoft SQL Server"},
				{Value: "oracle", Label: "Oracle"},
				{Value: "sqlite", Label: "SQLite"},
			},
		},
		{
			ID:      "level",
			Label:   "Nível (1-3)",
			Type:    "number",
			Default: "1",
			Help:    "Quanto maior, mais testes e payloads. Use 3 com autorização.",
			Min:     1,
			Max:     3,
		},
		{
			ID:      "risk",
			Label:   "Risco (1-3)",
			Type:    "number",
			Default: "1",
			Help:    "Risco 3 usa payloads mais agressivos (ex.: OOB).",
			Min:     1,
			Max:     3,
		},
		{
			ID:      "threads",
			Label:   "Threads",
			Type:    "number",
			Default: "4",
			Min:     1,
		},
		{
			ID:    "dump",
			Label: "Dump do banco de dados",
			Type:  "boolean",
			Help:  "Adiciona --dump para extrair o conteúdo do banco.",
		},
		{
			ID:      "batch",
			Label:   "Modo batch (sem perguntas)",
			Type:    "boolean",
			Default: "true",
		},
	},
}

var metasploitTool = &Tool{
	ID:          "metasploit",
	Name:        "Metasploit",
	Icon:        "💀",
	Description: "Framework de exploração completo. Gera payloads com msfvenom e scripts de resource (msfconsole).",
	Category:    CatExploit,
	Risk:        "high",
	Install:     "sudo apt install metasploit-framework",
	Docs:        "https://docs.metasploit.com/",
	Tags:        []string{"exploração", "payload", "meterpreter", "c2"},
	Questions: []Question{
		{
			ID:      "platform",
			Label:   "Plataforma do alvo",
			Type:    "select",
			Default: "linux",
			Options: []Option{
				{Value: "linux", Label: "Linux"},
				{Value: "windows", Label: "Windows"},
				{Value: "android", Label: "Android"},
			},
		},
		{
			ID:      "listener",
			Label:   "Tipo de payload",
			Type:    "select",
			Default: "reverse_tcp",
			Options: []Option{
				{Value: "reverse_tcp", Label: "Reverse TCP"},
				{Value: "reverse_https", Label: "Reverse HTTPS (melhor para firewalls)"},
				{Value: "bind_tcp", Label: "Bind TCP"},
			},
		},
		{
			ID:          "lhost",
			Label:       "LHOST (seu IP de atacante)",
			Type:        "text",
			Placeholder: "10.10.14.5",
			Required:    true,
		},
		{
			ID:      "lport",
			Label:   "LPORT",
			Type:    "number",
			Default: "4444",
		},
		{
			ID:          "rtarget",
			Label:       "Alvo remoto (RHOSTS, opcional)",
			Type:        "text",
			Placeholder: "192.168.1.50",
		},
		{
			ID:          "module",
			Label:       "Módulo de exploit (opcional)",
			Type:        "text",
			Placeholder: "exploit/windows/smb/ms17_010_eternalblue",
			Help:        "Se vazio, usa exploit/multi/handler para receber o payload gerado.",
		},
	},
}

var hydraTool = &Tool{
	ID:          "hydra",
	Name:        "Hydra",
	Icon:        "🔑",
	Description: "Força bruta de logins em serviços de rede (SSH, FTP, HTTP, SMB, RDP e mais).",
	Category:    CatCreds,
	Risk:        "medium",
	Install:     "sudo apt install hydra",
	Docs:        "https://github.com/vanhauser-thc/thc-hydra",
	Tags:        []string{"credenciais", "força bruta", "senhas"},
	Questions: []Question{
		{
			ID:      "service",
			Label:   "Serviço alvo",
			Type:    "select",
			Default: "ssh",
			Options: []Option{
				{Value: "ssh", Label: "SSH"},
				{Value: "ftp", Label: "FTP"},
				{Value: "http-post-form", Label: "HTTP POST form (login web)"},
				{Value: "smb", Label: "SMB"},
				{Value: "rdp", Label: "RDP"},
				{Value: "mysql", Label: "MySQL"},
				{Value: "telnet", Label: "Telnet"},
			},
		},
		{
			ID:          "target",
			Label:       "IP / host alvo",
			Type:        "text",
			Placeholder: "192.168.1.10",
			Required:    true,
		},
		{
			ID:          "login",
			Label:       "Usuário único (opcional)",
			Type:        "text",
			Placeholder: "admin",
			Help:        "Usa -l. Prefira isto sobre lista quando souber o usuário.",
		},
		{
			ID:          "userlist",
			Label:       "Lista de usuários (opcional, usa -L)",
			Type:        "text",
			Placeholder: "/usr/share/wordlists/usernames.txt",
		},
		{
			ID:          "passlist",
			Label:       "Lista de senhas",
			Type:        "text",
			Placeholder: "/usr/share/wordlists/rockyou.txt",
			Default:     "/usr/share/wordlists/rockyou.txt",
			Required:    true,
		},
		{
			ID:          "formdata",
			Label:       "Dados do form (apenas HTTP POST)",
			Type:        "text",
			Placeholder: "username=^USER^&password=^PASS^:Login failed",
			Help:        "Formato: postdata:marca de falha. ^USER^ e ^PASS^ são placeholders.",
		},
		{
			ID:          "port",
			Label:       "Porta (opcional)",
			Type:        "number",
			Placeholder: "2222",
			Min:         1,
			Max:         65535,
		},
		{
			ID:      "threads",
			Label:   "Threads",
			Type:    "number",
			Default: "16",
			Min:     1,
		},
		{
			ID:    "verbose",
			Label: "Modo verboso (-vV)",
			Type:  "boolean",
		},
	},
}

var hashcatTool = &Tool{
	ID:          "hashcat",
	Name:        "Hashcat",
	Icon:        "⚡",
	Description: "O quebrador de senhas mais rápido do mundo, com aceleração por GPU.",
	Category:    CatCreds,
	Risk:        "low",
	Install:     "sudo apt install hashcat",
	Docs:        "https://hashcat.net/wiki/",
	Tags:        []string{"credenciais", "hashes", "gpu", "senhas"},
	Questions: []Question{
		{
			ID:      "hashType",
			Label:   "Tipo de hash",
			Type:    "select",
			Default: "0",
			Options: []Option{
				{Value: "0", Label: "MD5"},
				{Value: "100", Label: "SHA1"},
				{Value: "1400", Label: "SHA256"},
				{Value: "1000", Label: "NTLM (Windows)"},
				{Value: "1800", Label: "sha512crypt (Linux)"},
				{Value: "3200", Label: "bcrypt ($2b$)"},
				{Value: "22000", Label: "WPA-PBKDF2-PMKID (Wi-Fi)"},
			},
		},
		{
			ID:      "attack",
			Label:   "Modo de ataque",
			Type:    "select",
			Default: "0",
			Options: []Option{
				{Value: "0", Label: "Dicionário (-a 0)"},
				{Value: "1", Label: "Combinador (-a 1)"},
				{Value: "3", Label: "Máscara / força bruta (-a 3)"},
			},
		},
		{
			ID:          "hashfile",
			Label:       "Arquivo de hashes",
			Type:        "text",
			Placeholder: "/tmp/hashes.txt",
			Required:    true,
		},
		{
			ID:          "wordlist",
			Label:       "Wordlist (ataque 0 e 1)",
			Type:        "text",
			Placeholder: "/usr/share/wordlists/rockyou.txt",
			Default:     "/usr/share/wordlists/rockyou.txt",
		},
		{
			ID:          "mask",
			Label:       "Máscara (ataque 3)",
			Type:        "text",
			Placeholder: "?u?l?l?l?d?d?d?d",
			Help:        "?l a-z, ?u A-Z, ?d 0-9, ?s símbolos, ?a tudo.",
		},
		{
			ID:          "rules",
			Label:       "Regras (opcional)",
			Type:        "text",
			Placeholder: "/usr/share/hashcat/rules/best64.rule",
			Help:        "Adiciona -r. Ex.: best64.rule, d3ad0ne.rule.",
		},
		{
			ID:      "device",
			Label:   "Dispositivo",
			Type:    "select",
			Default: "gpu",
			Options: []Option{
				{Value: "gpu", Label: "GPU (padrão)"},
				{Value: "cpu", Label: "CPU (-D 1)"},
			},
		},
	},
}

var aircrackTool = &Tool{
	ID:          "aircrack-ng",
	Name:        "Aircrack-ng",
	Icon:        "📶",
	Description: "Suite completa para auditoria de redes Wi-Fi: captura de handshake WPA/WPA2 e quebra da senha.",
	Category:    CatWireless,
	Risk:        "medium",
	Install:     "sudo apt install aircrack-ng",
	Docs:        "https://www.aircrack-ng.org/",
	Tags:        []string{"wi-fi", "wpa", "wireless", "handshake"},
	Questions: []Question{
		{
			ID:          "iface",
			Label:       "Interface wireless",
			Type:        "text",
			Placeholder: "wlan0",
			Required:    true,
		},
		{
			ID:          "bssid",
			Label:       "BSSID do AP (opcional)",
			Type:        "text",
			Placeholder: "AA:BB:CC:DD:EE:FF",
			Help:        "Deixe vazio para capturar todas as redes visíveis.",
		},
		{
			ID:          "channel",
			Label:       "Canal (opcional)",
			Type:        "number",
			Placeholder: "6",
			Min:         1,
			Max:         165,
		},
		{
			ID:      "capture",
			Label:   "Prefixo do arquivo de captura",
			Type:    "text",
			Default: "capture",
		},
		{
			ID:          "wordlist",
			Label:       "Wordlist para quebra",
			Type:        "text",
			Placeholder: "/usr/share/wordlists/rockyou.txt",
			Default:     "/usr/share/wordlists/rockyou.txt",
			Required:    true,
		},
	},
}

var tcpdumpTool = &Tool{
	ID:          "tcpdump",
	Name:        "tcpdump",
	Icon:        "📡",
	Description: "Captura e análise de tráfego de rede em tempo real via CLI.",
	Category:    CatNetwork,
	Risk:        "low",
	Install:     "sudo apt install tcpdump",
	Docs:        "https://www.tcpdump.org/manpages/tcpdump.1.html",
	Tags:        []string{"rede", "captura", "pacotes", "pcap"},
	Questions: []Question{
		{
			ID:          "iface",
			Label:       "Interface",
			Type:        "text",
			Placeholder: "eth0",
			Default:     "any",
			Help:        "Use 'any' para capturar em todas as interfaces.",
		},
		{
			ID:          "host",
			Label:       "Host alvo (opcional)",
			Type:        "text",
			Placeholder: "10.10.10.5",
		},
		{
			ID:          "port",
			Label:       "Porta (opcional)",
			Type:        "number",
			Placeholder: "443",
			Min:         1,
			Max:         65535,
		},
		{
			ID:          "filter",
			Label:       "Filtro BPF extra (opcional)",
			Type:        "text",
			Placeholder: "tcp port 80 or udp port 53",
			Help:        "Expressão de filtro do Berkeley Packet Filter.",
		},
		{
			ID:          "count",
			Label:       "Número de pacotes (opcional)",
			Type:        "number",
			Placeholder: "1000",
			Min:         1,
		},
		{
			ID:          "output",
			Label:       "Arquivo de saída .pcap (opcional)",
			Type:        "text",
			Placeholder: "captura.pcap",
			Help:        "Grava com -w para análise posterior no Wireshark.",
		},
		{
			ID:    "verbose",
			Label: "Modo verboso (-v)",
			Type:  "boolean",
		},
	},
}

var impacketTool = &Tool{
	ID:          "impacket",
	Name:        "Impacket / NetExec",
	Icon:        "🪟",
	Description: "Coleção de scripts Python para interação com protocolos Windows (SMB, WMI, MSSQL) e movimento lateral em AD.",
	Category:    CatPostExploit,
	Risk:        "high",
	Install:     "sudo apt install impacket-scripts && pipx install netexec",
	Docs:        "https://github.com/fortra/impacket",
	Tags:        []string{"windows", "active directory", "movimento lateral", "smb"},
	Questions: []Question{
		{
			ID:      "tool",
			Label:   "Script / módulo",
			Type:    "select",
			Default: "psexec",
			Options: []Option{
				{Value: "psexec", Label: "psexec.py — execução remota (SMB)"},
				{Value: "wmiexec", Label: "wmiexec.py — execução via WMI"},
				{Value: "smbclient", Label: "smbclient.py — listar/enviar arquivos"},
				{Value: "secretsdump", Label: "secretsdump.py — dump de hashes"},
				{Value: "mssql", Label: "mssqlclient.py — console SQL"},
				{Value: "netexec", Label: "netexec smb — enumeração de hosts AD"},
			},
		},
		{
			ID:          "target",
			Label:       "IP / host alvo",
			Type:        "text",
			Placeholder: "192.168.1.20",
			Required:    true,
		},
		{
			ID:          "domain",
			Label:       "Domínio (opcional)",
			Type:        "text",
			Placeholder: "corp.local",
		},
		{
			ID:          "user",
			Label:       "Usuário",
			Type:        "text",
			Placeholder: "administrator",
		},
		{
			ID:          "password",
			Label:       "Senha (ou use hash)",
			Type:        "text",
			Placeholder: "P@ssw0rd",
		},
		{
			ID:          "hash",
			Label:       "Hash NT/NTLM (opcional)",
			Type:        "text",
			Placeholder: "aad3b435b51404eeaad3b435b51404ee",
			Help:        "Usa -hashes para pass-the-hash.",
		},
		{
			ID:          "command",
			Label:       "Comando a executar (psexec/wmiexec)",
			Type:        "text",
			Placeholder: "whoami",
		},
	},
}

var mimikatzTool = &Tool{
	ID:          "mimikatz",
	Name:        "Mimikatz",
	Icon:        "👤",
	Description: "Extrai credenciais de memória (LSASS), hashes do SAM/LSA, tickets Kerberos e executa pass-the-hash/ticket no Windows.",
	Category:    CatPostExploit,
	Risk:        "high",
	Install:     "Baixe o binário em https://github.com/gentilkiwi/mimikatz/releases",
	Docs:        "https://github.com/gentilkiwi/mimikatz",
	Tags:        []string{"windows", "credenciais", "lsass", "kerberos", "pass-the-hash"},
	Questions: []Question{
		{
			ID:      "module",
			Label:   "Ação principal",
			Type:    "select",
			Default: "logonpasswords",
			Options: []Option{
				{Value: "logonpasswords", Label: "Dump de credenciais em memória (sekurlsa::logonpasswords)"},
				{Value: "sam", Label: "Dump do SAM local (lsadump::sam)"},
				{Value: "lsa", Label: "Segredos do LSA (lsadump::lsa /patch)"},
				{Value: "dcsync", Label: "DCSync — sincronizar diretório (lsadump::dcsync)"},
				{Value: "pth", Label: "Pass-the-Hash (sekurlsa::pth)"},
				{Value: "ptt", Label: "Pass-the-Ticket (kerberos::ptt)"},
			},
		},
		{
			ID:      "elevate",
			Label:   "Elevar para SYSTEM (token::elevate)",
			Type:    "boolean",
			Default: "true",
			Help:    "Necessário para acesso ao SAM/LSA. Requer execução como Administrador.",
		},
		{
			ID:          "domain",
			Label:       "Domínio (DCSync e Pass-the-Hash)",
			Type:        "text",
			Placeholder: "corp.local",
		},
		{
			ID:          "user",
			Label:       "Usuário alvo (DCSync / Pass-the-Hash)",
			Type:        "text",
			Placeholder: "Administrator",
		},
		{
			ID:          "hash",
			Label:       "Hash NT/NTLM (Pass-the-Hash)",
			Type:        "text",
			Placeholder: "aad3b435b51404eeaad3b435b51404ee",
			Help:        "Usa /ntlm:<hash> em sekurlsa::pth.",
		},
		{
			ID:          "ticket",
			Label:       "Caminho do ticket .kirbi (Pass-the-Ticket)",
			Type:        "text",
			Placeholder: "C:\\Users\\user\\ticket.kirbi",
		},
	},
}

var caidoTool = &Tool{
	ID:          "caido",
	Name:        "Caido",
	Icon:        "🕸️",
	Description: "Kit de auditoria web moderno (alternativa ao Burp): proxy de interceptação, repeater, workflows e suporte a GraphQL.",
	Category:    CatWeb,
	Risk:        "low",
	Install:     "sudo apt install caido (Kali) ou baixe em https://caido.io/download",
	Docs:        "https://docs.caido.io/",
	Tags:        []string{"web", "proxy", "interceptação", "burp"},
	Questions: []Question{
		{
			ID:      "deploy",
			Label:   "Forma de execução",
			Type:    "select",
			Default: "cli",
			Options: []Option{
				{Value: "cli", Label: "Caido CLI (caido-cli) — GUI no navegador"},
				{Value: "docker", Label: "Docker (caido/caido)"},
				{Value: "desktop", Label: "Aplicativo desktop"},
			},
		},
		{
			ID:      "port",
			Label:   "Porta do proxy",
			Type:    "number",
			Default: "8080",
			Help:    "Configure o proxy do navegador para esta porta.",
		},
		{
			ID:          "datadir",
			Label:       "Volume de dados (Docker, opcional)",
			Type:        "text",
			Placeholder: "/home/user/caido-data",
			Help:        "Monta o diretório para persistir os projetos do Caido.",
		},
		{
			ID:    "noopen",
			Label: "Não abrir o navegador automaticamente",
			Type:  "boolean",
			Help:  "Adiciona --no-open (útil em servidores headless).",
		},
	},
}

var bloodhoundTool = &Tool{
	ID:          "bloodhound",
	Name:        "BloodHound",
	Icon:        "🐺",
	Description: "Mapeamento e análise de caminhos de ataque em Active Directory usando grafos (coleta + visualização).",
	Category:    CatPostExploit,
	Risk:        "medium",
	Install:     "sudo apt install bloodhound bloodhound-python",
	Docs:        "https://bloodhound.readthedocs.io/",
	Tags:        []string{"windows", "active directory", "enumeração", "grafos"},
	Questions: []Question{
		{
			ID:      "collector",
			Label:   "Coletor de dados",
			Type:    "select",
			Default: "bloodhound-python",
			Options: []Option{
				{Value: "bloodhound-python", Label: "bloodhound-python (Linux)"},
				{Value: "sharphound", Label: "SharpHound.exe (Windows)"},
			},
		},
		{
			ID:          "user",
			Label:       "Usuário",
			Type:        "text",
			Placeholder: "svc_probe",
			Required:    true,
		},
		{
			ID:          "password",
			Label:       "Senha (ou use hash)",
			Type:        "text",
			Placeholder: "P@ssw0rd",
		},
		{
			ID:          "hash",
			Label:       "Hash NT (opcional, -k/--hashes)",
			Type:        "text",
			Placeholder: "aad3b435b51404eeaad3b435b51404ee",
		},
		{
			ID:          "domain",
			Label:       "Domínio",
			Type:        "text",
			Placeholder: "corp.local",
			Required:    true,
		},
		{
			ID:          "ns",
			Label:       "Nameserver / DC (opcional)",
			Type:        "text",
			Placeholder: "192.168.1.10",
			Help:        "Usa -ns <ip> para resolução de nomes no bloodhound-python.",
		},
		{
			ID:      "method",
			Label:   "Método de coleta",
			Type:    "select",
			Default: "All",
			Options: []Option{
				{Value: "All", Label: "All — tudo (recomendado)"},
				{Value: "Default", Label: "Default"},
				{Value: "Session", Label: "Session — sessões ativas (barulhento)"},
				{Value: "ACL", Label: "ACL — permissões de objetos"},
			},
		},
		{
			ID:      "neo4j",
			Label:   "Iniciar Neo4j (banco de grafos)",
			Type:    "boolean",
			Default: "true",
		},
	},
}

var fluxionTool = &Tool{
	ID:          "fluxion",
	Name:        "Fluxion",
	Icon:        "🕶️",
	Description: "Ataque de engenharia social WPA/WPA2 (evil twin): captura de handshake e portal cativo que engana o usuário.",
	Category:    CatWireless,
	Risk:        "high",
	Install:     "git clone https://github.com/FluxionNetwork/fluxion.git",
	Docs:        "https://github.com/FluxionNetwork/fluxion",
	Tags:        []string{"wi-fi", "wpa", "evil twin", "engenharia social", "portal cativo"},
	Questions: []Question{
		{
			ID:          "iface",
			Label:       "Interface wireless",
			Type:        "text",
			Placeholder: "wlan0",
			Required:    true,
			Help:        "Deve suportar modo monitor e packet injection.",
		},
		{
			ID:      "attack",
			Label:   "Tipo de ataque",
			Type:    "select",
			Default: "handshake",
			Options: []Option{
				{Value: "handshake", Label: "Handshake Snooper — capturar handshake"},
				{Value: "captive", Label: "Captive Portal — portal falso (evil twin)"},
			},
		},
		{
			ID:          "essid",
			Label:       "SSID da rede alvo (opcional)",
			Type:        "text",
			Placeholder: "Casa_Wifi",
		},
		{
			ID:          "bssid",
			Label:       "BSSID do AP (opcional)",
			Type:        "text",
			Placeholder: "AA:BB:CC:DD:EE:FF",
		},
		{
			ID:          "channel",
			Label:       "Canal (opcional)",
			Type:        "number",
			Placeholder: "6",
			Min:         1,
			Max:         165,
		},
		{
			ID:      "lang",
			Label:   "Idioma da interface",
			Type:    "select",
			Default: "pt",
			Options: []Option{
				{Value: "pt", Label: "Português"},
				{Value: "en", Label: "Inglês"},
				{Value: "es", Label: "Espanhol"},
			},
		},
	},
}

var ffufTool = &Tool{
	ID:          "ffuf",
	Name:        "ffuf",
	Icon:        "🔥",
	Description: "Fuzzing web extremamente rápido para descoberta de diretórios, virtual hosts e pontos de injeção usando wordlists.",
	Category:    CatWeb,
	Risk:        "low",
	Install:     "sudo apt install ffuf",
	Docs:        "https://github.com/ffuf/ffuf",
	Tags:        []string{"web", "fuzzing", "diretórios", "vhost", "brute force"},
	Questions: []Question{
		{
			ID:      "mode",
			Label:   "Modo de fuzzing",
			Type:    "select",
			Default: "dir",
			Options: []Option{
				{Value: "dir", Label: "Diretórios (-u http://alvo/FUZZ)"},
				{Value: "vhost", Label: "Virtual hosts (Header Host: FUZZ.domínio)"},
				{Value: "fuzz", Label: "Parâmetros / pontos de injeção (FUZZ na URL)"},
			},
		},
		{
			ID:          "url",
			Label:       "URL alvo (com a palavra-chave FUZZ)",
			Type:        "text",
			Placeholder: "http://example.com/FUZZ",
			Required:    true,
			Help:        "Inclua FUZZ onde a wordlist será injetada (ex.: http://alvo/FUZZ ou http://alvo/?q=FUZZ).",
		},
		{
			ID:          "domain",
			Label:       "Domínio base (modo vhost)",
			Type:        "text",
			Placeholder: "example.com",
			Help:        "Monta o header Host: FUZZ.<domínio>. Obrigatório no modo vhost.",
		},
		{
			ID:          "wordlist",
			Label:       "Wordlist",
			Type:        "text",
			Placeholder: "/usr/share/seclists/Discovery/Web-Content/directory-list-2.3-medium.txt",
			Default:     "/usr/share/seclists/Discovery/Web-Content/directory-list-2.3-medium.txt",
		},
		{
			ID:          "extensions",
			Label:       "Extensões (modo dir, opcional)",
			Type:        "text",
			Placeholder: "php,html,txt",
			Help:        "Usa -e. Separe com vírgulas.",
		},
		{
			ID:      "threads",
			Label:   "Threads",
			Type:    "number",
			Default: "40",
			Min:     1,
		},
		{
			ID:          "fc",
			Label:       "Filtrar status codes (opcional)",
			Type:        "text",
			Placeholder: "200,301,302,403",
			Help:        "Usa -fc para ocultar respostas com esses códigos.",
		},
		{
			ID:          "output",
			Label:       "Arquivo de saída JSON (opcional)",
			Type:        "text",
			Placeholder: "resultado.json",
		},
	},
}

var msfvenomTool = &Tool{
	ID:          "msfvenom",
	Name:        "msfvenom",
	Icon:        "🧬",
	Description: "Gerador de payloads do Metasploit: cria executáveis e raw payloads para Linux, Windows, Android e macOS.",
	Category:    CatExploit,
	Risk:        "high",
	Install:     "sudo apt install metasploit-framework",
	Docs:        "https://docs.metasploit.com/docs/using-metasploit/basics/generating-payloads.html",
	Tags:        []string{"payload", "exploração", "meterpreter", "evasão"},
	Questions: []Question{
		{
			ID:      "platform",
			Label:   "Plataforma do alvo",
			Type:    "select",
			Default: "linux",
			Options: []Option{
				{Value: "linux", Label: "Linux"},
				{Value: "windows", Label: "Windows"},
				{Value: "android", Label: "Android"},
				{Value: "macos", Label: "macOS"},
			},
		},
		{
			ID:      "listener",
			Label:   "Tipo de payload",
			Type:    "select",
			Default: "reverse_tcp",
			Options: []Option{
				{Value: "reverse_tcp", Label: "Reverse TCP"},
				{Value: "reverse_https", Label: "Reverse HTTPS (melhor para firewalls)"},
				{Value: "bind_tcp", Label: "Bind TCP"},
			},
		},
		{
			ID:          "lhost",
			Label:       "LHOST (seu IP de atacante)",
			Type:        "text",
			Placeholder: "10.10.14.5",
			Required:    true,
		},
		{
			ID:      "lport",
			Label:   "LPORT",
			Type:    "number",
			Default: "4444",
		},
		{
			ID:      "format",
			Label:   "Formato de saída",
			Type:    "select",
			Default: "auto",
			Options: []Option{
				{Value: "auto", Label: "Automático (elf/exe/apk/macho)"},
				{Value: "elf", Label: "ELF (Linux)"},
				{Value: "exe", Label: "EXE (Windows)"},
				{Value: "exe-service", Label: "EXE como serviço Windows"},
				{Value: "apk", Label: "APK (Android)"},
				{Value: "macho", Label: "Macho (macOS)"},
				{Value: "raw", Label: "Raw"},
				{Value: "ps1", Label: "PowerShell (PS1)"},
			},
		},
		{
			ID:          "encoder",
			Label:       "Encoder (opcional)",
			Type:        "text",
			Placeholder: "x86/shikata_ga_nai -i 5",
			Help:        "Ex.: x86/shikata_ga_nai com -i 5. Não garante evasão de AV.",
		},
	},
}

var masscanTool = &Tool{
	ID:          "masscan",
	Name:        "masscan",
	Icon:        "💥",
	Description: "Scanner de portas extremamente rápido (milhões de pacotes/s) para redes e ranges de IP grandes.",
	Category:    CatNetwork,
	Risk:        "medium",
	Install:     "sudo apt install masscan",
	Docs:        "https://github.com/robertdavidgraham/masscan",
	Tags:        []string{"rede", "scanner", "portas", "internet"},
	Questions: []Question{
		{
			ID:          "targets",
			Label:       "Alvos (IP, range ou rede)",
			Type:        "text",
			Placeholder: "10.0.0.0/8",
			Required:    true,
			Help:        "Aceita CIDR (10.0.0.0/8), range (10.0.0.1-10.0.0.254) ou IP único.",
		},
		{
			ID:          "ports",
			Label:       "Portas",
			Type:        "text",
			Placeholder: "80,443,8080,1-65535",
			Default:     "80,443",
			Help:        "Lista separada por vírgulas ou intervalo. Use 1-65535 para todas.",
		},
		{
			ID:      "rate",
			Label:   "Taxa de pacotes/s",
			Type:    "number",
			Default: "1000",
			Help:    "--rate. Comece baixo e aumente conforme a rede aguentar.",
			Min:     1,
		},
		{
			ID:          "iface",
			Label:       "Interface (opcional)",
			Type:        "text",
			Placeholder: "eth0",
			Help:        "Usa -e. Necessário em máquinas com múltiplas interfaces.",
		},
		{
			ID:    "output",
			Label: "Formatos de saída",
			Type:  "multi",
			Options: []Option{
				{Value: "normal", Label: "Texto normal (-oL)"},
				{Value: "xml", Label: "XML (-oX)"},
				{Value: "json", Label: "JSON (-oJ)"},
				{Value: "grepable", Label: "Grepable (-oG)"},
			},
		},
	},
}

var beefTool = &Tool{
	ID:          "beef",
	Name:        "beEF",
	Icon:        "🐝",
	Description: "Browser Exploitation Framework: fisga navegadores via hook.js e permite keylogging, screenshots, pivoting e mais.",
	Category:    CatExploit,
	Risk:        "high",
	Install:     "sudo apt install beef-xss (Kali) ou clone https://github.com/beefproject/beef",
	Docs:        "https://beefproject.com/",
	Tags:        []string{"exploração", "browser", "hook", "xss", "c2"},
	Questions: []Question{
		{
			ID:      "install",
			Label:   "Forma de instalação",
			Type:    "select",
			Default: "apt",
			Options: []Option{
				{Value: "apt", Label: "Pacote Kali (beef-xss)"},
				{Value: "source", Label: "Clone do GitHub (./beef)"},
			},
		},
		{
			ID:      "port",
			Label:   "Porta HTTP",
			Type:    "number",
			Default: "3000",
		},
		{
			ID:          "hookHost",
			Label:       "Endereço acessível pelas vítimas",
			Type:        "text",
			Placeholder: "10.10.14.5",
			Required:    true,
			Help:        "IP do seu host que a vítima conseguirá alcançar (a interface acessível pela rede alvo).",
		},
	},
}

var kismetTool = &Tool{
	ID:          "kismet",
	Name:        "Kismet",
	Icon:        "📻",
	Description: "Detector de redes wireless e dispositivos por captura passiva de quadros 802.11.",
	Category:    CatWireless,
	Risk:        "low",
	Install:     "sudo apt install kismet",
	Docs:        "https://www.kismetwireless.net/",
	Tags:        []string{"wi-fi", "wireless", "detecção", "802.11"},
	Questions: []Question{
		{
			ID:          "iface",
			Label:       "Interface wireless",
			Type:        "text",
			Placeholder: "wlan0",
			Required:    true,
		},
		{
			ID:      "channelHop",
			Label:   "Channel hopping",
			Type:    "boolean",
			Default: "true",
			Help:    "Alterna canais automaticamente. Desative para fixar em um canal.",
		},
		{
			ID:          "logtitle",
			Label:       "Prefixo dos logs (opcional)",
			Type:        "text",
			Placeholder: "kismet",
			Help:        "Usa -t como título base dos arquivos de log.",
		},
	},
}

var xsserTool = &Tool{
	ID:          "xsser",
	Name:        "XSSer",
	Icon:        "🎯",
	Description: "Scanner de Cross-Site Scripting (XSS) que testa múltiplos vetores e engines de browsers em URLs web.",
	Category:    CatWeb,
	Risk:        "medium",
	Install:     "sudo apt install xsser",
	Docs:        "https://xsser.03c8.net/",
	Tags:        []string{"web", "xss", "injeção", "scanner"},
	Questions: []Question{
		{
			ID:          "url",
			Label:       "URL alvo (com ponto de injeção)",
			Type:        "text",
			Placeholder: "http://example.com/search.php?q=",
			Required:    true,
			Help:        "Termine com o ponto de injeção: http://alvo/page.php?param=",
		},
		{
			ID:      "auto",
			Label:   "Modo automático (--auto)",
			Type:    "boolean",
			Default: "true",
			Help:    "Testa automaticamente todos os vetores XSS disponíveis.",
		},
		{
			ID:          "cookie",
			Label:       "Cookie de sessão (opcional)",
			Type:        "text",
			Placeholder: "PHPSESSID=abc123",
			Help:        "Usa -c para requisições autenticadas.",
		},
		{
			ID:      "threads",
			Label:   "Threads",
			Type:    "number",
			Default: "8",
			Min:     1,
		},
		{
			ID:    "verbose",
			Label: "Modo verboso (-v)",
			Type:  "boolean",
		},
	},
}

var commixTool = &Tool{
	ID:          "commix",
	Name:        "Commix",
	Icon:        "💉",
	Description: "Exploração automatizada de injeção de comandos (OS command injection) em parâmetros HTTP.",
	Category:    CatWeb,
	Risk:        "high",
	Install:     "sudo apt install commix",
	Docs:        "https://github.com/commixproject/commix",
	Tags:        []string{"web", "injeção de comandos", "rce", "exploração"},
	Questions: []Question{
		{
			ID:          "url",
			Label:       "URL da requisição",
			Type:        "text",
			Placeholder: "http://example.com/ping.php?host=1.1.1.1",
			Required:    true,
		},
		{
			ID:      "method",
			Label:   "Método HTTP",
			Type:    "select",
			Default: "get",
			Options: []Option{
				{Value: "get", Label: "GET"},
				{Value: "post", Label: "POST"},
			},
		},
		{
			ID:          "data",
			Label:       "Corpo POST / parâmetros (opcional)",
			Type:        "text",
			Placeholder: "host=1.1.1.1",
			Help:        "Obrigatório quando o método é POST.",
		},
		{
			ID:      "level",
			Label:   "Nível (1-3)",
			Type:    "number",
			Default: "1",
			Help:    "Maior nível = mais payloads e técnicas.",
			Min:     1,
			Max:     3,
		},
		{
			ID:      "os",
			Label:   "Sistema operacional alvo",
			Type:    "select",
			Default: "auto",
			Options: []Option{
				{Value: "auto", Label: "Auto-detecção"},
				{Value: "unix", Label: "Unix/Linux"},
				{Value: "windows", Label: "Windows"},
			},
		},
		{
			ID:      "batch",
			Label:   "Modo batch (sem perguntas)",
			Type:    "boolean",
			Default: "true",
		},
		{
			ID:    "shell",
			Label: "Spawn de shell interativa (--os-shell)",
			Type:  "boolean",
			Help:  "Tenta abrir uma sessão de shell quando possível.",
		},
		{
			ID:    "randomAgent",
			Label: "User-Agent aleatório",
			Type:  "boolean",
			Help:  "Adiciona --random-agent.",
		},
	},
}

var sherlockTool = &Tool{
	ID:          "sherlock",
	Name:        "Sherlock",
	Icon:        "🕵️",
	Description: "OSINT: encontra perfis de um usuário em mais de 300 redes sociais e sites, buscando pelo nome de usuário.",
	Category:    CatRecon,
	Risk:        "low",
	Install:     "pip install sherlock-project",
	Docs:        "https://github.com/sherlock-project/sherlock",
	Tags:        []string{"osint", "usuários", "redes sociais", "reconhecimento"},
	Questions: []Question{
		{
			ID:          "usernames",
			Label:       "Nome(s) de usuário",
			Type:        "text",
			Placeholder: "joao.silva ou user1,user2",
			Required:    true,
			Help:        "Separe múltiplos nomes com vírgula.",
		},
		{
			ID:      "timeout",
			Label:   "Timeout por site (s)",
			Type:    "number",
			Default: "5",
			Min:     1,
		},
		{
			ID:      "printFound",
			Label:   "Mostrar apenas resultados encontrados",
			Type:    "boolean",
			Default: "true",
			Help:    "Adiciona --print-found.",
		},
		{
			ID:      "noColor",
			Label:   "Sem cores na saída",
			Type:    "boolean",
			Default: "true",
			Help:    "Adiciona --no-color.",
		},
		{
			ID:          "output",
			Label:       "Arquivo de saída (opcional)",
			Type:        "text",
			Placeholder: "resultado.txt",
			Help:        "Usa -o.",
		},
	},
}

var holeheTool = &Tool{
	ID:          "holehe",
	Name:        "Holehe",
	Icon:        "📧",
	Description: "OSINT: verifica se um endereço de e-mail está registrado em mais de 120 serviços e redes sociais.",
	Category:    CatRecon,
	Risk:        "low",
	Install:     "pip install holehe",
	Docs:        "https://github.com/megadose/holehe",
	Tags:        []string{"osint", "e-mail", "reconhecimento", "redes sociais"},
	Questions: []Question{
		{
			ID:          "email",
			Label:       "Endereço de e-mail",
			Type:        "text",
			Placeholder: "nome@example.com",
			Required:    true,
		},
		{
			ID:    "onlyUsed",
			Label: "Mostrar apenas contas existentes",
			Type:  "boolean",
			Help:  "Adiciona --only-used.",
		},
		{
			ID:      "noColor",
			Label:   "Sem cores na saída",
			Type:    "boolean",
			Default: "true",
			Help:    "Adiciona --no-color.",
		},
		{
			ID:          "csv",
			Label:       "Exportar CSV (opcional)",
			Type:        "text",
			Placeholder: "resultado.csv",
			Help:        "Usa -C <arquivo>.",
		},
	},
}

var rdnsTool = &Tool{
	ID:          "rdns",
	Name:        "Reverse Genie (rDNS)",
	Icon:        "🌐",
	Description: "Reverse DNS lookup em massa: mapeia IPs para nomes de host usando dig, dnsrecon ou nmap -sL.",
	Category:    CatRecon,
	Risk:        "low",
	Install:     "sudo apt install dnsutils dnsrecon nmap",
	Docs:        "https://github.com/darkoperator/dnsrecon",
	Tags:        []string{"reconhecimento", "dns", "rdns", "mapeamento"},
	Questions: []Question{
		{
			ID:          "target",
			Label:       "Alvo (IP, range ou rede)",
			Type:        "text",
			Placeholder: "10.0.0.1 ou 10.0.0.0/24",
			Required:    true,
			Help:        "Um IP para dig, ou um range/CIDR para dnsrecon e nmap -sL.",
		},
		{
			ID:      "method",
			Label:   "Ferramenta",
			Type:    "select",
			Default: "dig",
			Options: []Option{
				{Value: "dig", Label: "dig -x (IP único)"},
				{Value: "dnsrecon", Label: "dnsrecon -r (range/rede)"},
				{Value: "nmap", Label: "nmap -sL (range/rede, sem varredura)"},
			},
		},
		{
			ID:          "resolver",
			Label:       "Nameserver (opcional)",
			Type:        "text",
			Placeholder: "8.8.8.8",
			Help:        "Usa @server no dig e -n no dnsrecon.",
		},
		{
			ID:          "output",
			Label:       "Arquivo de saída (opcional)",
			Type:        "text",
			Placeholder: "rdns.csv",
			Help:        "dnsrecon usa -c (CSV); nmap usa -oN.",
		},
	},
}
