#!/usr/bin/env python3
"""
CupaX Analysis Agent - HTTP API Server

CRITICAL SECURITY WARNING:
This agent MUST be run in a completely isolated and sandboxed Windows VM!
The agent executes potentially malicious files and WILL be compromised.

Recommended setup:
1. Dedicated Windows VM (Windows 10/11)
2. Install Noriben and dependencies (Procmon, Python 3)
3. Network isolation: Only allow inbound from CupaX backend IP
4. Regular VM snapshot/revert after each analysis
5. NO internet access for the VM
"""

import argparse
import json
import logging
import os
import re
import shutil
import subprocess
import tempfile
import zipfile
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Any

from flask import Flask, request, jsonify

# Configuration from environment or defaults
NORIBEN_PATH = os.getenv("AGENT_NORIBEN_PATH", "Noriben.py")
PYTHON_PATH = os.getenv("AGENT_PYTHON_PATH", "python")  # Use "python" on Windows
ANALYSIS_TIMEOUT = int(os.getenv("AGENT_TIMEOUT", "300"))  # 5 minutes
WORK_DIR = os.getenv("AGENT_WORK_DIR", "./agent_work")
AGENT_PORT = int(os.getenv("AGENT_PORT", "9090"))
AGENT_HOST = os.getenv("AGENT_HOST", "0.0.0.0")  # Listen on all interfaces
UNZIP_TOOL = os.getenv("AGENT_UNZIP_TOOL", "7z")  # Options: "7z", "unzip", or "python"
UNZIP_PATH = os.getenv("AGENT_UNZIP_PATH", "7z")  # Path to unzip tool executable

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("cupax-agent")

# Flask app
app = Flask(__name__)


class NoribenParser:
    """Parser for Noriben output files"""

    @staticmethod
    def parse_txt_report(filepath: str) -> Dict[str, Any]:
        """Parse the Noriben .txt summary report"""
        try:
            with open(filepath, 'r', encoding='utf-8-sig', errors='ignore') as f:
                content = f.read()

            summary = {
                "execution_time": 0.0,
                "processing_time": 0.0,
                "analysis_time": 0.0,
                "processes_created": 0,
                "files_created": 0,
                "registry_modified": 0,
                "network_connections": 0
            }

            # Extract timing information
            exec_match = re.search(r'Execution time: ([\d.]+) seconds', content)
            if exec_match:
                summary["execution_time"] = float(exec_match.group(1))

            proc_match = re.search(r'Processing time: ([\d.]+) seconds', content)
            if proc_match:
                summary["processing_time"] = float(proc_match.group(1))

            anal_match = re.search(r'Analysis time: ([\d.]+) seconds', content)
            if anal_match:
                summary["analysis_time"] = float(anal_match.group(1))

            # Count events by section
            process_section = re.search(r'Processes Created:\n={15,}\n(.*?)(?:\n\n|\Z)', content, re.DOTALL)
            if process_section:
                summary["processes_created"] = len([l for l in process_section.group(1).split('\n') if l.strip()])

            file_section = re.search(r'File Activity:\n={15,}\n(.*?)(?:\n\n|\Z)', content, re.DOTALL)
            if file_section:
                summary["files_created"] = len([l for l in file_section.group(1).split('\n') if l.strip()])

            reg_section = re.search(r'Registry Activity:\n={15,}\n(.*?)(?:\n\n|\Z)', content, re.DOTALL)
            if reg_section:
                summary["registry_modified"] = len([l for l in reg_section.group(1).split('\n') if l.strip()])

            net_section = re.search(r'Network Traffic:\n={15,}\n(.*?)(?:\n\n|\Z)', content, re.DOTALL)
            if net_section:
                summary["network_connections"] = len([l for l in net_section.group(1).split('\n') if l.strip()])

            # Extract unique hosts
            hosts_section = re.search(r'Unique Hosts:\n={15,}\n(.*?)(?:\n\n|\Z)', content, re.DOTALL)
            unique_hosts = []
            if hosts_section:
                unique_hosts = [h.strip() for h in hosts_section.group(1).split('\n') if h.strip()]

            return {"summary": summary, "unique_hosts": unique_hosts}

        except Exception as e:
            logger.error(f"Failed to parse TXT report: {e}")
            return {"summary": {}, "unique_hosts": []}

    @staticmethod
    def parse_csv_timeline(filepath: str) -> Dict[str, List[Dict]]:
        """Parse the Noriben timeline CSV"""
        import csv

        events = {
            "process_activity": [],
            "file_system": [],
            "registry": [],
            "network": []
        }

        try:
            with open(filepath, 'r', encoding='utf-8-sig', errors='ignore') as f:
                reader = csv.reader(f)
                for row in reader:
                    if len(row) < 3:
                        continue

                    timestamp = row[0]
                    category = row[1]
                    operation = row[2]

                    if category == "Process":
                        events["process_activity"].append({
                            "timestamp": timestamp,
                            "process_name": row[3] if len(row) > 3 else "",
                            "pid": row[4] if len(row) > 4 else "",
                            "command_line": row[5] if len(row) > 5 else "",
                            "child_pid": row[6] if len(row) > 6 else ""
                        })

                    elif category == "File":
                        event = {
                            "timestamp": timestamp,
                            "operation": operation,
                            "process_name": row[3] if len(row) > 3 else "",
                            "pid": row[4] if len(row) > 4 else "",
                            "path": row[5] if len(row) > 5 else ""
                        }

                        if len(row) > 6:
                            event["hash_type"] = row[6] if row[6] else ""
                        if len(row) > 7:
                            event["hash"] = row[7] if row[7] else ""
                        if len(row) > 8:
                            event["yara_hits"] = row[8] if row[8] else ""
                        if len(row) > 9:
                            event["vt_hits"] = row[9] if row[9] else ""

                        if operation == "RenameFile" and len(row) > 6:
                            event["to_path"] = row[6]

                        events["file_system"].append(event)

                    elif category == "Registry":
                        event = {
                            "timestamp": timestamp,
                            "operation": operation,
                            "process_name": row[3] if len(row) > 3 else "",
                            "pid": row[4] if len(row) > 4 else "",
                            "path": row[5] if len(row) > 5 else "",
                            "data": row[6] if len(row) > 6 else ""
                        }
                        events["registry"].append(event)

                    elif category == "Network":
                        protocol = operation
                        direction = "Unknown"
                        if ' ' in operation:
                            parts = operation.split(' ', 1)
                            protocol = parts[0]
                            direction = parts[1] if len(parts) > 1 else "Unknown"

                        event = {
                            "timestamp": timestamp,
                            "protocol": protocol,
                            "direction": direction,
                            "process_name": row[3] if len(row) > 3 else "",
                            "pid": row[4] if len(row) > 4 else "",
                            "remote_addr": row[5] if len(row) > 5 else ""
                        }
                        events["network"].append(event)

            return events

        except Exception as e:
            logger.error(f"Failed to parse CSV timeline: {e}")
            return events


class MalwareAnalyzer:
    """Handles malware analysis with Noriben"""

    def __init__(self):
        self.work_dir = Path(WORK_DIR)
        self.work_dir.mkdir(exist_ok=True, parents=True)

    def extract_zip(self, zip_path: str, analysis_id: str, password: Optional[str] = None) -> Optional[str]:
        """Extract zip file and return path to executable"""
        try:
            # Use consistent naming: {analysis_id}_extracted
            extract_dir = self.work_dir / f"{analysis_id}_extracted"
            extract_dir.mkdir(exist_ok=True, parents=True)
            logger.info(f"Extracting to: {extract_dir}")

            # Use configured extraction tool
            if UNZIP_TOOL == "7z":
                success = self._extract_with_7z(zip_path, extract_dir, password)
            elif UNZIP_TOOL == "unzip":
                success = self._extract_with_unzip(zip_path, extract_dir, password)
            else:  # python (default fallback)
                success = self._extract_with_python(zip_path, extract_dir, password)

            if not success:
                logger.error("Failed to extract zip file")
                return None

            # Find executable file
            executables = []
            for ext in ['*.exe', '*.dll', '*.scr', '*.bat', '*.cmd', '*.ps1']:
                executables.extend(extract_dir.glob(ext))
                executables.extend(extract_dir.glob(f'**/{ext}'))

            if not executables:
                logger.error("No executable found in zip")
                return None

            # Return first executable
            exe_path = str(executables[0])
            logger.info(f"Extracted executable: {exe_path}")
            return exe_path

        except Exception as e:
            logger.error(f"Failed to extract zip: {e}")
            return None

    def _extract_with_7z(self, zip_path: str, extract_dir: Path, password: Optional[str] = None) -> bool:
        """Extract using 7-Zip (most robust, supports all formats)"""
        try:
            cmd = [UNZIP_PATH, "x", zip_path, f"-o{extract_dir}", "-y"]

            if password:
                cmd.append(f"-p{password}")
            else:
                # Try without password first
                cmd.append("-p")  # Empty password

            logger.info(f"Extracting with 7z: {' '.join(cmd[:4])}...")  # Don't log password

            result = subprocess.run(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                timeout=30
            )

            if result.returncode == 0:
                logger.info("7z extraction successful")
                return True
            else:
                # Try common passwords if first attempt failed
                if not password:
                    common_passwords = ['infected', 'malware', 'virus', 'password']
                    for pwd in common_passwords:
                        cmd_with_pwd = [UNZIP_PATH, "x", zip_path, f"-o{extract_dir}", "-y", f"-p{pwd}"]
                        result = subprocess.run(cmd_with_pwd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=30)
                        if result.returncode == 0:
                            logger.info(f"7z extraction successful with password: {pwd}")
                            return True

                logger.error(f"7z extraction failed: {result.stderr}")
                return False

        except Exception as e:
            logger.error(f"7z extraction error: {e}")
            return False

    def _extract_with_unzip(self, zip_path: str, extract_dir: Path, password: Optional[str] = None) -> bool:
        """Extract using unzip.exe (UnixUtils for Windows)"""
        try:
            cmd = [UNZIP_PATH, "-o", zip_path, "-d", str(extract_dir)]

            if password:
                cmd.extend(["-P", password])

            logger.info(f"Extracting with unzip: {' '.join(cmd[:4])}...")

            result = subprocess.run(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                timeout=30
            )

            if result.returncode == 0:
                logger.info("unzip extraction successful")
                return True
            else:
                # Try common passwords
                if not password:
                    common_passwords = ['infected', 'malware', 'virus', 'password']
                    for pwd in common_passwords:
                        cmd_with_pwd = [UNZIP_PATH, "-o", zip_path, "-d", str(extract_dir), "-P", pwd]
                        result = subprocess.run(cmd_with_pwd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=30)
                        if result.returncode == 0:
                            logger.info(f"unzip extraction successful with password: {pwd}")
                            return True

                logger.error(f"unzip extraction failed: {result.stderr}")
                return False

        except Exception as e:
            logger.error(f"unzip extraction error: {e}")
            return False

    def _extract_with_python(self, zip_path: str, extract_dir: Path, password: Optional[str] = None) -> bool:
        """Extract using Python's zipfile module (fallback, limited format support)"""
        try:
            with zipfile.ZipFile(zip_path, 'r') as zip_ref:
                if password:
                    # Try to extract with password
                    zip_ref.extractall(extract_dir, pwd=password.encode())
                    logger.info("Python zipfile extraction successful with password")
                    return True
                else:
                    # Try without password first
                    try:
                        zip_ref.extractall(extract_dir)
                        logger.info("Python zipfile extraction successful")
                        return True
                    except RuntimeError:
                        # Try common passwords
                        common_passwords = ['infected', 'malware', 'virus', 'password']
                        for pwd in common_passwords:
                            try:
                                zip_ref.extractall(extract_dir, pwd=pwd.encode())
                                logger.info(f"Python zipfile extraction successful with password: {pwd}")
                                return True
                            except RuntimeError:
                                continue

                        logger.error("Python zipfile: password required or unsupported compression")
                        return False

        except Exception as e:
            logger.error(f"Python zipfile extraction error: {e}")
            return False

    def run_noriben(self, sample_path: str, analysis_id: str) -> Dict[str, Any]:
        """Execute Noriben analysis"""
        try:
            logger.info(f"Starting Noriben analysis: {analysis_id}")

            # Create output directory
            output_dir = self.work_dir / analysis_id
            output_dir.mkdir(exist_ok=True, parents=True)

            # Convert paths to absolute paths
            abs_sample_path = os.path.abspath(sample_path)
            abs_noriben_path = os.path.abspath(NORIBEN_PATH)
            abs_output_dir = os.path.abspath(str(output_dir))

            # Verify sample file exists
            if not os.path.exists(abs_sample_path):
                logger.error(f"Sample file not found: {abs_sample_path}")
                return {
                    "success": False,
                    "error": f"Sample file not found: {abs_sample_path}"
                }

            cmd = [
                PYTHON_PATH,
                abs_noriben_path,
                "--cmd", abs_sample_path,
                "--timeout", str(ANALYSIS_TIMEOUT),
                "--headless",
                "--output", abs_output_dir
            ]

            logger.info(f"Executing: {' '.join(cmd)}")

            # Run Noriben with timeout
            logger.info(f"Sample file size: {os.path.getsize(abs_sample_path)} bytes")
            logger.info(f"Working directory: {os.getcwd()}")

            process = subprocess.run(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                timeout=ANALYSIS_TIMEOUT + 60,
                text=True,
                cwd=os.path.dirname(abs_noriben_path)  # Run from Noriben directory
            )

            logger.info(f"Noriben returned with code: {process.returncode}")
            if process.stdout:
                logger.info(f"STDOUT: {process.stdout[:500]}")  # Log first 500 chars

            if process.returncode != 0:
                logger.error(f"Noriben failed with code {process.returncode}")
                logger.error(f"STDERR: {process.stderr}")
                logger.error(f"STDOUT: {process.stdout}")
                return {
                    "success": False,
                    "error": f"Noriben execution failed (code {process.returncode}): {process.stderr}"
                }

            # Parse results
            report = self.parse_results(str(output_dir))
            if not report:
                return {
                    "success": False,
                    "error": "Failed to parse Noriben output"
                }

            logger.info("Analysis completed successfully")
            return {
                "success": True,
                "report": report
            }

        except subprocess.TimeoutExpired:
            logger.error("Noriben execution timed out")
            return {
                "success": False,
                "error": "Analysis timeout"
            }
        except Exception as e:
            logger.error(f"Analysis failed: {e}")
            return {
                "success": False,
                "error": str(e)
            }
        finally:
            # Cleanup sample and temp files
            try:
                if os.path.exists(sample_path):
                    os.remove(sample_path)
                    logger.info(f"Cleaned up sample: {sample_path}")
            except Exception as e:
                logger.warning(f"Failed to cleanup sample: {e}")

            # Cleanup extraction directory if it exists
            try:
                extract_dir = self.work_dir / f"{analysis_id}_extracted"
                if extract_dir.exists():
                    shutil.rmtree(extract_dir)
                    logger.info(f"Cleaned up extraction directory: {extract_dir}")
            except Exception as e:
                logger.warning(f"Failed to cleanup extraction directory: {e}")

    def parse_results(self, output_dir: str) -> Optional[Dict]:
        """Parse Noriben output files"""
        try:
            output_path = Path(output_dir)
            txt_files = list(output_path.glob("Noriben_*.txt"))
            csv_files = list(output_path.glob("Noriben_*_timeline.csv"))

            if not txt_files or not csv_files:
                logger.error("Noriben output files not found")
                return None

            txt_file = str(txt_files[0])
            csv_file = str(csv_files[0])

            logger.info(f"Parsing results from: {txt_file}, {csv_file}")

            # Parse both files
            txt_data = NoribenParser.parse_txt_report(txt_file)
            csv_data = NoribenParser.parse_csv_timeline(csv_file)

            # Combine into final report (send raw data to backend)
            report = {
                "summary": txt_data.get("summary", {}),
                "process_activity": csv_data.get("process_activity", []),
                "file_system": csv_data.get("file_system", []),
                "registry": csv_data.get("registry", []),
                "network": csv_data.get("network", []),
                "unique_hosts": txt_data.get("unique_hosts", [])
            }

            logger.info(f"Parsed {len(report['process_activity'])} process events, "
                       f"{len(report['file_system'])} file events, "
                       f"{len(report['registry'])} registry events, "
                       f"{len(report['network'])} network events")

            return report

        except Exception as e:
            logger.error(f"Failed to parse results: {e}")
            return None


# Initialize analyzer
analyzer = MalwareAnalyzer()


@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "noriben_path": NORIBEN_PATH,
        "work_dir": str(analyzer.work_dir),
        "timeout": ANALYSIS_TIMEOUT
    })


@app.route('/analyze', methods=['POST'])
def analyze():
    """
    Analyze a malware sample (SYNCHRONOUS VERSION)

    Request:
    - file: uploaded file (multipart/form-data)
    - analysis_id: unique analysis ID
    - password: (optional) zip password
    - is_zip: (optional) whether file is an infected zip

    Response:
    {
        "success": true/false,
        "report": {...} or "error": "..."
    }
    """
    try:
        # Get uploaded file
        if 'file' not in request.files:
            return jsonify({"success": False, "error": "No file uploaded"}), 400

        file = request.files['file']
        analysis_id = request.form.get('analysis_id')
        password = request.form.get('password')
        is_zip = request.form.get('is_zip', 'false').lower() == 'true'

        if not analysis_id:
            return jsonify({"success": False, "error": "analysis_id required"}), 400

        logger.info(f"Received analysis request: {analysis_id}")

        # Get original filename and extension
        original_filename = file.filename
        print(f"Original filename: {original_filename}")
        file_ext = os.path.splitext(original_filename)[1]  # Get extension like .exe, .dll, etc.
        # Save uploaded file with original extension preserved
        temp_file = analyzer.work_dir / f"{analysis_id}_upload{file_ext}"
        file.save(str(temp_file))
        logger.info(f"Saved uploaded file: {temp_file} (original: {original_filename})")

        # Handle zip extraction if needed
        if is_zip:
            logger.info(f"Extracting infected zip: {analysis_id}")
            sample_path = analyzer.extract_zip(str(temp_file), analysis_id, password)

            # Clean up zip file
            try:
                os.remove(str(temp_file))
            except:
                pass

            if not sample_path:
                return jsonify({
                    "success": False,
                    "error": "Failed to extract zip file"
                }), 400

        else:
            sample_path = str(temp_file)

        # Run analysis synchronously
        result = analyzer.run_noriben(sample_path, analysis_id)

        if result["success"]:
            return jsonify(result), 200
        else:
            return jsonify(result), 500

    except Exception as e:
        logger.error(f"Analysis request failed: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500


@app.route('/cleanup/<analysis_id>', methods=['DELETE'])
def cleanup(analysis_id):
    """Clean up analysis artifacts"""
    try:
        output_dir = analyzer.work_dir / analysis_id
        if output_dir.exists():
            shutil.rmtree(output_dir)

        return jsonify({"success": True}), 200
    except Exception as e:
        logger.error(f"Cleanup failed: {e}")
        return jsonify({"success": False, "error": str(e)}), 500


def main():
    parser = argparse.ArgumentParser(description='CupaX Analysis Agent')
    parser.add_argument('--host', default=AGENT_HOST, help='Host to bind to')
    parser.add_argument('--port', type=int, default=AGENT_PORT, help='Port to bind to')
    args = parser.parse_args()

    logger.info("="*60)
    logger.info("CupaX Analysis Agent Starting")
    logger.info("="*60)
    logger.info("SECURITY WARNING: This agent executes malware!")
    logger.info("Ensure you are running in an isolated VM!")
    logger.info("="*60)
    logger.info(f"Noriben path: {NORIBEN_PATH}")
    logger.info(f"Python path: {PYTHON_PATH}")
    logger.info(f"Analysis timeout: {ANALYSIS_TIMEOUT}s")
    logger.info(f"Work directory: {analyzer.work_dir}")
    logger.info(f"ZIP extraction tool: {UNZIP_TOOL}")
    logger.info(f"ZIP extraction path: {UNZIP_PATH}")
    logger.info(f"Listening on: {args.host}:{args.port}")
    logger.info("="*60)

    app.run(host=args.host, port=args.port, debug=False)


if __name__ == "__main__":
    main()
