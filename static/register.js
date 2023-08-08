// register.js

document.getElementById('registerForm').addEventListener('submit', function(event) {
    event.preventDefault(); // 阻止表单默认提交行为
    // 获取用户名、密码和昵称
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const name = document.getElementById('nickname').value;

    // 发送注册请求
    fetch('http://127.0.0.1:8888/api/storage/v0/register', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, password, name })
    })
        .then(response => response.json())
        .then(data => {
            // 注册成功，保存 token 到 Web Storage
            // localStorage.setItem('token', data.token);
            alert("注册成功")
            // 重定向到 index.html
            window.location.href = 'index.html';
        })
        .catch(error => {
            console.error('Registration failed:', error);
        });
});