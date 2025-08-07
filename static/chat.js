let ws;

function connect() {
    const name = document.getElementById("nameInput").value;
    if (!name) {
        alert("请输入昵称！");
        return;
    }

    ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = function () {
        ws.send(name); // 发送昵称作为第一条消息
    };

    ws.onmessage = function (event) {
        const chatBox = document.getElementById("chat");
        chatBox.innerHTML += event.data + "<br/>";
        chatBox.scrollTop = chatBox.scrollHeight;
    };
}

function sendMessage() {
    const msg = document.getElementById("msgInput").value;
    if (ws && msg) {
        ws.send(msg);
        document.getElementById("msgInput").value = "";
    }
}
