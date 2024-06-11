from flask import Flask, request, jsonify
import subprocess

app = Flask(__name__)

def get_vmid_by_ip(ip):
    command = f"cat $(grep -Ril 'ip={ip}' /etc/pve/lxc/*.conf) | grep rootfs | cut -d':' -f3 | cut -d'/' -f1 | cut -d'-' -f2"
    vmid = subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT, universal_newlines=True).strip()
    return vmid

def get_vmid_by_hostname(hostname):
    command = f"cat $(grep -Ril 'hostname={hostname}' /etc/pve/lxc/*.conf) | grep rootfs | cut -d':' -f3 | cut -d'/' -f1 | cut -d'-' -f2"
    vmid = subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT, universal_newlines=True).strip()
    return vmid

@app.route('/ssh', methods=['POST'])
def deploy():
    data = request.get_json()
    ip = data.get('ip')
    hostname = data.get('hostname')
    key = data.get('key')

    
    if not key:
        return jsonify({"error": "No key provided"}), 400

    try:
        if ip != None:
            vmid = get_vmid_by_ip(ip)
        elif hostname != None:
            vmid = get_vmid_by_hostname(hostname)
        else:
            return jsonify({"error": "IP address or hostname not provided"}), 400

        if vmid == "":
            return jsonify({"error": "VMID not found"}), 404
        
        command = f"pct exec {vmid} -- bash -c 'echo \"{key}\" >> /root/.ssh/authorized_keys'"
        output = subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT, universal_newlines=True)

        return jsonify({"output": output})
    except subprocess.CalledProcessError as e:
        return jsonify({"error": str(e), "output": e.output}), 500

@app.route('/ssh/vmid', methods=['POST'])
def deploy_vmid():
    data = request.get_json()
    vmid = data.get('vmid')
    key = data.get('key')

    if not vmid or not key:
        return jsonify({"error": "VMID or key not provided"}), 400

    try:
        command = f"pct exec {vmid} -- bash -c 'echo \"{key}\" >> /root/.ssh/authorized_keys'"
        output = subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT, universal_newlines=True)

        return jsonify({"output": output})
    except subprocess.CalledProcessError as e:
        return jsonify({"error": str(e), "output": e.output}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5252)