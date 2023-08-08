// login.js

document.getElementById('loginForm').addEventListener('submit', function(event) {
    event.preventDefault(); // 阻止表单默认提交行为

    // 获取用户名和密码
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    // 发送登录请求
    fetch('http://127.0.0.1:8888/api/storage/v0/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, password })
    })
        .then(response => response.json())
        .then(data => {
            // 登录成功，保存 token 到 Web Storage
            alert(data.data)
            localStorage.setItem('x-token', data.data);
            alert("登陆成功")
            // 重定向到 index.html
            window.location.href = 'index.html';
        })
        .catch(error => {
            console.error('Login failed:', error);
        });
});