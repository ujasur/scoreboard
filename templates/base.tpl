[[ template "root" .]]
[[ define "base" ]]
<!DOCTYPE html>
<html lang="en">
<head lang="en">
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="ie=edge">
  <title>ScoreBoard - [[block "title" .]] [[end]]</title>
  <link rel="icon" type="image/png" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAABKUlEQVQ4T62Tu2oCURCGvxUJioUoWGknXioVKxXBxk5bnyaQLi/jC4hgIpaCIoog2ghWImxQ8IIruOEwLCQsu24uP5xu5pt//sNoZrv1AjwDT/xMBvCqme3W9RfN1ihDAUzXweEwaBocDmDaS50Bfj9UqxCLCV/Xod+3QZwBmQzkcjAYCKBWg7d3+NC/GXYGFIuQTMJkAvu9ADodOJ89AuJxqFTgfpf953PYbm1xuYdouVBtCrBYeARks6DSHw6hUIBUSsLr9WSdL7I7iESgXof1GkYj+cJmEwIBmE5htXoASKchn4flEmYz8Pmg0RCAAiqwq4NEAsplMAzYbCAalXe5QLcLt9sDgLJcKoECWdrtYDyG49FjiKosFIJgEE4nme6gfzmmP53zJ3cBhNODL9CBAAAAAElFTkSuQmCC">  
  <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css" integrity="sha384-Vkoo8x4CGsO3+Hhxv8T/Q5PaXtkKtu6ug5TOeNV6gBiFeWPGFN9MuhOf23Q9Ifjh" crossorigin="anonymous">
  <link rel="stylesheet" href="/static/base.css" crossorigin="anonymous">
  [[block "style" .]] [[end]]
  <script src="https://code.jquery.com/jquery-3.4.1.min.js" integrity="sha256-CSXorXvZcTkaix6Yvo6HppcZGetbYMGWSFlBw8HfCJo=" crossorigin="anonymous"></script>
  <script src="/static/base.js"></script>
</head>
<body>
  [[block "nav" .]]
  <nav class="navbar navbar-expand-lg">
    <a class="navbar-brand team-text" href="/">
      [[ .Team ]]
    </a>
    <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false"
      aria-label="Toggle navigation">
      <span class="navbar-toggler-icon"></span>
    </button>
    <div class="collapse navbar-collapse" id="navbarNav">
      <ul class="navbar-nav mr-auto">
        <li class="nav-item [[if eq .Name "session.html" ]] active [[end]]">
          <a class="nav-link" href="/">Session
            <span class="sr-only">(current)</span>
          </a>
        </li>
        <li class="nav-item [[if eq .Name "links.html" ]] active [[end]]">
          <a class="nav-link" href="/ui/links">Links</a>
        </li>
        <li class="nav-item [[if eq .Name "users.html" ]] active [[end]]">
          <a class="nav-link" href="/ui/users">Users</a>
        </li>        
        <li class="nav-item [[if eq .Name "docs.html" ]] active [[end]]">
          <a class="nav-link" href="/ui/docs">Docs</a>
        </li>        
      </ul>
      <a id="LogoutBtn" href="#" onclick="logout()" class="btn btn-outline-secondary d-none">Logout</a>
    </div>
  </nav>
  [[end]]
  <div class="view">
    [[block "content" .]][[end]]    
  </div>
  [[ block "js" .]] [[end]]

  <div class="text-center mt-3">
    <small  class="d-md-block px-3 py-2" style="color: #5e6375">[[ .Version ]]</small>
  </div>
</body>
</html>
[[ end ]]