package usecase

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
)

// ThreatSignaturesDetector detects atomic threat labels across Security, Sysmon, System, and Application logs.
type ThreatSignaturesDetector struct{}

func NewThreatSignaturesDetector() *ThreatSignaturesDetector {
	return &ThreatSignaturesDetector{}
}

// DetectAlerts runs dual-inspection over an incoming SecurityEvent to identify threat alerts.
func (d *ThreatSignaturesDetector) DetectAlerts(evt *domain.SecurityEvent) []*domain.Alert {
	if evt == nil {
		return nil
	}
	var alerts []*domain.Alert

	host, account, ip := extractEntities(evt)
	eventID := extractEventID(evt)
	sourceClass := strings.ToLower(evt.Source)

	// Determine log source channel
	switch {
	case isSecurityLog(sourceClass, eventID, evt):
		alerts = append(alerts, d.detectSecurityAlerts(evt, eventID, host, account, ip)...)
	case isSysmonLog(sourceClass, eventID, evt):
		alerts = append(alerts, d.detectSysmonAlerts(evt, eventID, host, account, ip)...)
	case isSystemLog(sourceClass, eventID, evt):
		alerts = append(alerts, d.detectSystemAlerts(evt, eventID, host, account, ip)...)
	case isApplicationLog(sourceClass, eventID, evt):
		alerts = append(alerts, d.detectApplicationAlerts(evt, eventID, host, account, ip)...)
	default:
		// Fallback inspection across all signatures if source class is generic (e.g. "windows", "winevent", "log_upload")
		alerts = append(alerts, d.detectSecurityAlerts(evt, eventID, host, account, ip)...)
		alerts = append(alerts, d.detectSysmonAlerts(evt, eventID, host, account, ip)...)
		alerts = append(alerts, d.detectSystemAlerts(evt, eventID, host, account, ip)...)
	}

	for _, a := range alerts {
		a.OrganizationID = evt.OrganizationID
	}

	return alerts
}

// 1. Security Log Signatures
func (d *ThreatSignaturesDetector) detectSecurityAlerts(evt *domain.SecurityEvent, eventID string, host, account, ip string) []*domain.Alert {
	var alerts []*domain.Alert
	detectedAt := evt.OccurredAt
	if detectedAt.IsZero() {
		detectedAt = time.Now()
	}

	switch eventID {
	case "4625":
		logonType := getSignaturePayloadInt(evt, "LogonType", "logon_type")
		// EID 4625 LogonType 3 (Network logon)
		alerts = append(alerts, &domain.Alert{
			ID:            uuid.New(),
			ThreatLabel:   "Brute_Force_Failed_Login",
			MITRE:         "T1110.001",
			LogSource:     "Security",
			EventID:       "4625",
			EntityHost:    host,
			EntityAccount: account,
			EntityIP:      ip,
			RawEvent:      evt,
			Context:       map[string]interface{}{"logon_type": logonType},
			DetectedAt:    detectedAt,
		})

	case "4624":
		logonType := getSignaturePayloadInt(evt, "LogonType", "logon_type")
		authPackage := strings.ToUpper(getSignaturePayloadString(evt, "AuthenticationPackageName", "auth_package"))
		switch logonType {
		case 3:
			if strings.Contains(authPackage, "NTLM") {
				alerts = append(alerts, &domain.Alert{
					ID:            uuid.New(),
					ThreatLabel:   "Lateral_Movement_SMB",
					MITRE:         "T1021.002",
					LogSource:     "Security",
					EventID:       "4624",
					EntityHost:    host,
					EntityAccount: account,
					EntityIP:      ip,
					RawEvent:      evt,
					Context:       map[string]interface{}{"logon_type": 3, "auth_package": authPackage},
					DetectedAt:    detectedAt,
				})
			} else {
				alerts = append(alerts, &domain.Alert{
					ID:            uuid.New(),
					ThreatLabel:   "Brute_Force_Successful_Auth",
					MITRE:         "T1078.003",
					LogSource:     "Security",
					EventID:       "4624",
					EntityHost:    host,
					EntityAccount: account,
					EntityIP:      ip,
					RawEvent:      evt,
					Context:       map[string]interface{}{"logon_type": 3},
					DetectedAt:    detectedAt,
				})
			}
		case 10:
			// RDP logon
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Lateral_Movement_RDP",
				MITRE:         "T1021.001",
				LogSource:     "Security",
				EventID:       "4624",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"logon_type": 10},
				DetectedAt:    detectedAt,
			})
		case 9:
			// NewCredentials
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Lateral_Movement_NewCredentials",
				MITRE:         "T1021",
				LogSource:     "Security",
				EventID:       "4624",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"logon_type": 9},
				DetectedAt:    detectedAt,
			})
		}

	case "4672":
		privs := strings.ToLower(getSignaturePayloadString(evt, "PrivilegeList", "privileges"))
		if strings.Contains(privs, "sedebugprivilege") || strings.Contains(privs, "seimpersonateprivilege") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Privileged_Account_Logon",
				MITRE:         "T1078",
				LogSource:     "Security",
				EventID:       "4672",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"privileges": privs},
				DetectedAt:    detectedAt,
			})
		}

	case "4740":
		alerts = append(alerts, &domain.Alert{
			ID:            uuid.New(),
			ThreatLabel:   "Account_Lockout",
			MITRE:         "T1110.001",
			LogSource:     "Security",
			EventID:       "4740",
			EntityHost:    host,
			EntityAccount: account,
			EntityIP:      ip,
			RawEvent:      evt,
			DetectedAt:    detectedAt,
		})

	case "4720":
		alerts = append(alerts, &domain.Alert{
			ID:            uuid.New(),
			ThreatLabel:   "New_User_Account_Created",
			MITRE:         "T1136.001",
			LogSource:     "Security",
			EventID:       "4720",
			EntityHost:    host,
			EntityAccount: account,
			EntityIP:      ip,
			RawEvent:      evt,
			DetectedAt:    detectedAt,
		})

	case "4732":
		groupSid := getSignaturePayloadString(evt, "TargetSid", "GroupSid", "TargetUserName")
		if strings.Contains(groupSid, "S-1-5-32-544") || strings.Contains(strings.ToLower(groupSid), "administrator") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "User_Added_To_Administrators",
				MITRE:         "T1098",
				LogSource:     "Security",
				EventID:       "4732",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"group": groupSid},
				DetectedAt:    detectedAt,
			})
		}

	case "4726":
		alerts = append(alerts, &domain.Alert{
			ID:            uuid.New(),
			ThreatLabel:   "User_Account_Deleted",
			MITRE:         "T1070",
			LogSource:     "Security",
			EventID:       "4726",
			EntityHost:    host,
			EntityAccount: account,
			EntityIP:      ip,
			RawEvent:      evt,
			DetectedAt:    detectedAt,
		})

	case "4688":
		cmdLine := strings.ToLower(getSignaturePayloadString(evt, "CommandLine", "command_line", "ProcessCommandLine"))
		newProcess := strings.ToLower(getSignaturePayloadString(evt, "NewProcessName", "process_name", "Image"))

		// Check Recon Process Tools
		reconTools := []string{"whoami", "net.exe", "ipconfig", "tasklist", "reg.exe", "wmic", "systeminfo", "nltest"}
		for _, tool := range reconTools {
			if strings.Contains(newProcess, tool) || strings.Contains(cmdLine, tool) {
				alerts = append(alerts, &domain.Alert{
					ID:            uuid.New(),
					ThreatLabel:   "Recon_Process_Chain",
					MITRE:         "T1087",
					LogSource:     "Security",
					EventID:       "4688",
					EntityHost:    host,
					EntityAccount: account,
					EntityIP:      ip,
					RawEvent:      evt,
					Context:       map[string]interface{}{"command_line": cmdLine, "tool": tool},
					DetectedAt:    detectedAt,
				})
				break
			}
		}

		// Check LOLBins
		if (strings.Contains(cmdLine, "certutil") && strings.Contains(cmdLine, "-decode")) ||
			strings.Contains(cmdLine, "rundll32") ||
			strings.Contains(cmdLine, "mshta") ||
			strings.Contains(cmdLine, "regsvr32") ||
			(strings.Contains(cmdLine, "powershell") && strings.Contains(cmdLine, "-encodedcommand")) {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "LOLBin_Execution",
				MITRE:         "T1059",
				LogSource:     "Security",
				EventID:       "4688",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"command_line": cmdLine},
				DetectedAt:    detectedAt,
			})
		}
	}

	return alerts
}

// 2. Sysmon Log Signatures
func (d *ThreatSignaturesDetector) detectSysmonAlerts(evt *domain.SecurityEvent, eventID string, host, account, ip string) []*domain.Alert {
	var alerts []*domain.Alert
	detectedAt := evt.OccurredAt
	if detectedAt.IsZero() {
		detectedAt = time.Now()
	}

	switch eventID {
	case "1":
		// Sysmon Process Create
		parentImage := strings.ToLower(getSignaturePayloadString(evt, "ParentImage", "parent_image"))
		cmdLine := strings.ToLower(getSignaturePayloadString(evt, "CommandLine", "command_line"))
		image := strings.ToLower(getSignaturePayloadString(evt, "Image", "image"))

		// Recon process parent-child anomaly
		if (strings.Contains(parentImage, "powershell") && strings.Contains(cmdLine, "cmd.exe")) ||
			(strings.Contains(parentImage, "wmic.exe") && strings.Contains(cmdLine, "powershell")) {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Recon_Process_Chain",
				MITRE:         "T1059.001",
				LogSource:     "Sysmon",
				EventID:       "1",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"parent_image": parentImage, "command_line": cmdLine},
				DetectedAt:    detectedAt,
			})
		}

		// LOLBin in Sysmon
		if (strings.Contains(cmdLine, "certutil") && strings.Contains(cmdLine, "-decode")) ||
			strings.Contains(cmdLine, "mshta") ||
			strings.Contains(cmdLine, "rundll32") ||
			strings.Contains(cmdLine, "regsvr32") ||
			(strings.Contains(image, "powershell") && strings.Contains(cmdLine, "-encodedcommand")) {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "LOLBin_Execution",
				MITRE:         "T1059",
				LogSource:     "Sysmon",
				EventID:       "1",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"command_line": cmdLine},
				DetectedAt:    detectedAt,
			})
		}

	case "3":
		// Sysmon Network Connection
		destPort := getSignaturePayloadInt(evt, "DestinationPort", "destination_port", "dst_port")
		destIP := getSignaturePayloadString(evt, "DestinationIp", "destination_ip", "dst_ip")
		isIPv4 := strings.Count(destIP, ".") == 3

		if destPort == 4444 {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Reverse_Shell_Connection",
				MITRE:         "T1059",
				LogSource:     "Sysmon",
				EventID:       "3",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      destIP,
				RawEvent:      evt,
				Context:       map[string]interface{}{"destination_port": 4444, "dst_ip": destIP},
				DetectedAt:    detectedAt,
			})
		} else if destPort == 443 && isIPv4 && !isPrivateIP(destIP) {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Outbound_C2_Connection",
				MITRE:         "T1071.001",
				LogSource:     "Sysmon",
				EventID:       "3",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      destIP,
				RawEvent:      evt,
				Context:       map[string]interface{}{"destination_port": 443, "dst_ip": destIP},
				DetectedAt:    detectedAt,
			})
		} else if destPort == 445 {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Lateral_Movement_SMB",
				MITRE:         "T1021.002",
				LogSource:     "Sysmon",
				EventID:       "3",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      destIP,
				RawEvent:      evt,
				Context:       map[string]interface{}{"destination_port": 445, "dst_ip": destIP},
				DetectedAt:    detectedAt,
			})
		}

	case "11":
		// Sysmon File Create
		targetFilename := strings.ToLower(getSignaturePayloadString(evt, "TargetFilename", "target_filename", "file_path"))
		if strings.Contains(targetFilename, "mimikatz") || strings.Contains(targetFilename, "mimi.exe") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Mimikatz_Dropped",
				MITRE:         "T1003.001",
				LogSource:     "Sysmon",
				EventID:       "11",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"file": targetFilename},
				DetectedAt:    detectedAt,
			})
		} else if (strings.Contains(targetFilename, "\\temp\\") || strings.Contains(targetFilename, "\\users\\public\\")) &&
			(strings.HasSuffix(targetFilename, ".exe") || strings.HasSuffix(targetFilename, ".dll")) {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Executable_Dropped_Temp",
				MITRE:         "T1105",
				LogSource:     "Sysmon",
				EventID:       "11",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"file": targetFilename},
				DetectedAt:    detectedAt,
			})
		} else if strings.Contains(targetFilename, "\\startup\\") && strings.HasSuffix(targetFilename, ".lnk") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Persistence_Startup_LNK",
				MITRE:         "T1547.001",
				LogSource:     "Sysmon",
				EventID:       "11",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"file": targetFilename},
				DetectedAt:    detectedAt,
			})
		}

	case "13":
		// Sysmon Registry Event
		targetObject := strings.ToLower(getSignaturePayloadString(evt, "TargetObject", "target_object", "registry_key"))
		details := strings.ToLower(getSignaturePayloadString(evt, "Details", "details", "value"))

		if strings.Contains(targetObject, "currentversion\\run") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Persistence_RunKey",
				MITRE:         "T1547.001",
				LogSource:     "Sysmon",
				EventID:       "13",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"registry_key": targetObject, "details": details},
				DetectedAt:    detectedAt,
			})
		} else if strings.Contains(targetObject, "image file execution options\\sethc.exe") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Persistence_StickyKeys_IFEO",
				MITRE:         "T1546.008",
				LogSource:     "Sysmon",
				EventID:       "13",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"registry_key": targetObject},
				DetectedAt:    detectedAt,
			})
		} else if strings.Contains(targetObject, "services\\windefend") && (strings.Contains(details, "start=4") || strings.Contains(details, "dword:00000004")) {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Defense_Disabled_Defender",
				MITRE:         "T1562.001",
				LogSource:     "Sysmon",
				EventID:       "13",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"registry_key": targetObject, "details": details},
				DetectedAt:    detectedAt,
			})
		}

	case "22":
		// Sysmon DNS Query
		queryName := strings.ToLower(getSignaturePayloadString(evt, "QueryName", "query_name", "domain"))
		if strings.Contains(queryName, "ngrok") || strings.Contains(queryName, "duckdns") ||
			strings.Contains(queryName, "no-ip") || strings.Contains(queryName, "c2-backup") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "DynDNS_C2_Resolution",
				MITRE:         "T1071.004",
				LogSource:     "Sysmon",
				EventID:       "22",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"query_name": queryName},
				DetectedAt:    detectedAt,
			})
		}
	}

	return alerts
}

// 3. System Log Signatures
func (d *ThreatSignaturesDetector) detectSystemAlerts(evt *domain.SecurityEvent, eventID string, host, account, ip string) []*domain.Alert {
	var alerts []*domain.Alert
	detectedAt := evt.OccurredAt
	if detectedAt.IsZero() {
		detectedAt = time.Now()
	}

	switch eventID {
	case "7045":
		// Service Installed
		imagePath := strings.ToLower(getSignaturePayloadString(evt, "ImagePath", "image_path", "service_file_name"))
		serviceName := getSignaturePayloadString(evt, "ServiceName", "service_name")
		if strings.Contains(imagePath, "\\users\\public\\") || strings.Contains(imagePath, "\\temp\\") ||
			strings.Contains(imagePath, "cmd.exe /c") || strings.Contains(imagePath, "powershell") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Persistence_Malicious_Service",
				MITRE:         "T1543.003",
				LogSource:     "System",
				EventID:       "7045",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"service_name": serviceName, "image_path": imagePath},
				DetectedAt:    detectedAt,
			})
		}

	case "7036":
		// Service State Change
		serviceName := strings.ToLower(getSignaturePayloadString(evt, "ServiceName", "param1", "service_name"))
		state := strings.ToLower(getSignaturePayloadString(evt, "State", "param2", "state"))
		if strings.Contains(state, "stopped") && (strings.Contains(serviceName, "windefend") ||
			strings.Contains(serviceName, "sense") || strings.Contains(serviceName, "mpssvc") ||
			strings.Contains(serviceName, "wuauserv")) {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Defense_Disabled_Service_Stopped",
				MITRE:         "T1562.001",
				LogSource:     "System",
				EventID:       "7036",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"service_name": serviceName, "state": state},
				DetectedAt:    detectedAt,
			})
		}

	case "41":
		// Kernel power crash / bugcheck
		alerts = append(alerts, &domain.Alert{
			ID:            uuid.New(),
			ThreatLabel:   "System_Crash_AntiForensics",
			MITRE:         "T1562",
			LogSource:     "System",
			EventID:       "41",
			EntityHost:    host,
			EntityAccount: account,
			EntityIP:      ip,
			RawEvent:      evt,
			DetectedAt:    detectedAt,
		})

	case "1":
		// System time changed
		alerts = append(alerts, &domain.Alert{
			ID:            uuid.New(),
			ThreatLabel:   "System_Time_Changed",
			MITRE:         "T1070.006",
			LogSource:     "System",
			EventID:       "1",
			EntityHost:    host,
			EntityAccount: account,
			EntityIP:      ip,
			RawEvent:      evt,
			DetectedAt:    detectedAt,
		})
	}

	return alerts
}

// 4. Application Log Signatures
func (d *ThreatSignaturesDetector) detectApplicationAlerts(evt *domain.SecurityEvent, eventID string, host, account, ip string) []*domain.Alert {
	var alerts []*domain.Alert
	detectedAt := evt.OccurredAt
	if detectedAt.IsZero() {
		detectedAt = time.Now()
	}

	switch eventID {
	case "1116", "1117", "1118", "1119":
		// Windows Defender Threat Detection / Remediation Failed
		threatName := getSignaturePayloadString(evt, "ThreatName", "threat_name")
		remediationErr := getSignaturePayloadString(evt, "ErrorCode", "remediation_error")
		if eventID == "1119" || strings.Contains(strings.ToLower(threatName), "failed") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Defender_Remediation_Failed",
				MITRE:         "T1562.001",
				LogSource:     "Application",
				EventID:       eventID,
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"threat_name": threatName, "error": remediationErr},
				DetectedAt:    detectedAt,
			})
		}

	case "4104":
		// PowerShell ScriptBlock Log
		scriptBlock := strings.ToLower(getSignaturePayloadString(evt, "ScriptBlockText", "script_block", "message"))
		if strings.Contains(scriptBlock, "sekurlsa") || strings.Contains(scriptBlock, "mimikatz") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "PowerShell_Mimikatz_Command",
				MITRE:         "T1003.001",
				LogSource:     "Application",
				EventID:       "4104",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"script_snippet": truncateString(scriptBlock, 200)},
				DetectedAt:    detectedAt,
			})
		}
		if strings.Contains(scriptBlock, "shadowcopy.delete") || strings.Contains(scriptBlock, "vssadmin delete shadows") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Shadow_Copy_Destruction",
				MITRE:         "T1490",
				LogSource:     "Application",
				EventID:       "4104",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"script_snippet": truncateString(scriptBlock, 200)},
				DetectedAt:    detectedAt,
			})
		}
		if strings.Contains(scriptBlock, "stop-service windefend") || strings.Contains(scriptBlock, "set-mppreference -disablerealtimemonitoring $true") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Defense_Disabled_PowerShell",
				MITRE:         "T1562.001",
				LogSource:     "Application",
				EventID:       "4104",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"script_snippet": truncateString(scriptBlock, 200)},
				DetectedAt:    detectedAt,
			})
		}
		// Data collection / exfiltration loop in PowerShell
		if strings.Contains(scriptBlock, "get-childitem") && (strings.Contains(scriptBlock, ".docx") || strings.Contains(scriptBlock, ".pdf") || strings.Contains(scriptBlock, ".xlsx")) {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Data_Collection_Files",
				MITRE:         "T1005",
				LogSource:     "Application",
				EventID:       "4104",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"script_snippet": truncateString(scriptBlock, 200)},
				DetectedAt:    detectedAt,
			})
		} else if strings.Contains(scriptBlock, "compress-archive") || strings.Contains(scriptBlock, "7z a") || strings.Contains(scriptBlock, "tar -czf") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Data_Archive_Compression",
				MITRE:         "T1560.001",
				LogSource:     "Application",
				EventID:       "4104",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"script_snippet": truncateString(scriptBlock, 200)},
				DetectedAt:    detectedAt,
			})
		} else if strings.Contains(scriptBlock, "invoke-webrequest") || strings.Contains(scriptBlock, "uploadfile") || strings.Contains(scriptBlock, "webclient.upload") {
			alerts = append(alerts, &domain.Alert{
				ID:            uuid.New(),
				ThreatLabel:   "Data_Exfiltration_Upload",
				MITRE:         "T1048",
				LogSource:     "Application",
				EventID:       "4104",
				EntityHost:    host,
				EntityAccount: account,
				EntityIP:      ip,
				RawEvent:      evt,
				Context:       map[string]interface{}{"script_snippet": truncateString(scriptBlock, 200)},
				DetectedAt:    detectedAt,
			})
		}
	}

	return alerts
}

// Helper Functions for Dual-Inspection & Entity Extraction

func extractEntities(evt *domain.SecurityEvent) (host, account, ip string) {
	if evt == nil {
		return "unknown-host", "", ""
	}
	// 1. Host
	host = getSignaturePayloadString(evt, "Computer", "host", "hostname", "HostName", "ComputerName")
	if host == "" {
		host = "unknown-host"
	}

	// 2. Account
	if evt.ActorUsername != nil && *evt.ActorUsername != "" {
		account = *evt.ActorUsername
	} else {
		account = getSignaturePayloadString(evt, "TargetUserName", "user", "username", "AccountName", "User", "SubjectUserName")
	}

	// 3. IP
	if evt.IPAddress != nil && *evt.IPAddress != "" {
		ip = *evt.IPAddress
	} else {
		ip = getSignaturePayloadString(evt, "IpAddress", "ip_address", "source_ip", "c-ip", "ClientIP", "SourceAddress")
	}

	return host, account, ip
}

func extractEventID(evt *domain.SecurityEvent) string {
	if evt == nil {
		return ""
	}
	if id := getSignaturePayloadString(evt, "EventID", "event_id", "EventId"); id != "" {
		return id
	}
	if strings.Contains(evt.EventType, ":") {
		parts := strings.Split(evt.EventType, ":")
		if len(parts) > 1 && len(parts[1]) >= 1 {
			return parts[1]
		}
	}
	if _, err := strconv.Atoi(evt.EventType); err == nil {
		return evt.EventType
	}
	return ""
}

// Dual-Inspection payload lookups: Checks top-level / normalized payload first, then raw payload case-insensitively.
func getSignaturePayloadString(evt *domain.SecurityEvent, keys ...string) string {
	if evt == nil {
		return ""
	}
	// 1. Check NormalizedPayload
	if evt.NormalizedPayload != nil {
		for _, k := range keys {
			if v, ok := evt.NormalizedPayload[k]; ok && v != nil {
				return fmt.Sprintf("%v", v)
			}
		}
	}

	// 2. Fallback: Search RawPayload
	if evt.RawPayload != nil {
		// Exact match
		for _, k := range keys {
			if v, ok := evt.RawPayload[k]; ok && v != nil {
				return fmt.Sprintf("%v", v)
			}
		}
		// Case-insensitive match
		for k, v := range evt.RawPayload {
			for _, target := range keys {
				if strings.EqualFold(k, target) && v != nil {
					return fmt.Sprintf("%v", v)
				}
			}
		}
	}

	return ""
}

func getSignaturePayloadInt(evt *domain.SecurityEvent, keys ...string) int {
	str := getSignaturePayloadString(evt, keys...)
	if str == "" {
		return 0
	}
	if val, err := strconv.Atoi(str); err == nil {
		return val
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return int(f)
	}
	return 0
}

func isSecurityLog(source string, eventID string, evt *domain.SecurityEvent) bool {
	if strings.Contains(source, "security") {
		return true
	}
	switch eventID {
	case "4624", "4625", "4672", "4740", "4720", "4732", "4726", "4688":
		return true
	}
	return false
}

func isSysmonLog(source string, eventID string, evt *domain.SecurityEvent) bool {
	if strings.Contains(source, "sysmon") {
		return true
	}
	return false
}

func isSystemLog(source string, eventID string, evt *domain.SecurityEvent) bool {
	if strings.Contains(source, "system") {
		return true
	}
	switch eventID {
	case "7045", "7036":
		return true
	}
	return false
}

func isApplicationLog(source string, eventID string, evt *domain.SecurityEvent) bool {
	if strings.Contains(source, "application") {
		return true
	}
	switch eventID {
	case "1116", "1117", "1118", "1119", "4104", "1000":
		return true
	}
	return false
}

func isPrivateIP(ip string) bool {
	if strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "127.") {
		return true
	}
	if strings.HasPrefix(ip, "172.") {
		parts := strings.Split(ip, ".")
		if len(parts) > 1 {
			if n, err := strconv.Atoi(parts[1]); err == nil && n >= 16 && n <= 31 {
				return true
			}
		}
	}
	return false
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

