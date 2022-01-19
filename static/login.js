var optionsTemplate = Handlebars.compile(`
<h5>Your username</h5>
<div class="btn-group-vertical d-block" role="group" aria-label="Button group with nested dropdown">
{{#each users}}
  <button type="button" class="btn btn-block btn-light btn-lg user-btn" data-user="{{this.name}}">{{this.name}}</button>
{{/each}}
`);

var passcodeTemplate = Handlebars.compile(`
<div id="LoginForm">
  <div class="mb-3 clearfix">
    <div class="form-group">
      <label class="h5">{{name}}</label>
      <input id="Passcode" type="password" class="form-control form-control-lg" placeholder="Enter passcode" data-username="{{name}}">
    </div>    
    <div>
      <button class="btn btn-fixed-2 btn-outline-secondary float-left" type="button" style="cursor: pointer;" onclick="javascript:location.reload();">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16" fill="#fff">
          <path fill-rule="evenodd" d="M7.78 12.53a.75.75 0 01-1.06 0L2.47 8.28a.75.75 0 010-1.06l4.25-4.25a.75.75 0 011.06 1.06L4.81 7h7.44a.75.75 0 010 1.5H4.81l2.97 2.97a.75.75 0 010 1.06z"></path>
        </svg>
      </button>
      <button id="LoginBtn" class="btn btn-fixed-2 btn-primary float-right" type="button">Login</button>
    </div>    
  </div>  
  <div id="LoginError" class="alert alert-danger d-none mt-2"></div>
</div>           
`);

var LoginPage = {
  init(users) {
    this.users = users;
    this.root = $('#LoginPage');
    this.root.on('click', '.user-btn', (event) => {
      var name = event.target.dataset.user;
      if (event.target.dataset.user == 'observer') {
        window.localStorage.setItem("user", JSON.stringify({ name }));
        window.location = '/';
        return;
      }
      var user = this.users.find((u) => u.name === name);
      this.passcodeView(user);
    });

    this.root.keypress(function (event) {
      var keycode = (event.keyCode ? event.keyCode : event.which);
      if (keycode == '13' && $("#Passcode").length) {
        attemptLogin();
      }
    });

    this.root.on('click', '#LoginBtn', attemptLogin);

    function attemptLogin() {
      var passEl = $("#Passcode");
      var name = passEl.data('username');
      var passcode = passEl.val();
      api.userAuth({ name, passcode }, (data, statusText, res) => {
        var token = res.getResponseHeader('authorization');
        var role = atob(token).split(',')[2];
        window.localStorage.setItem("user", JSON.stringify({ name, token, role }));
        window.location = '/';
      },
        function (res) {
          var msg = res.error.error;
          $('#LoginError').removeClass('d-none').text(msg);
        });
    }

    this.root.html(optionsTemplate({ users: this.users }));
  },

  passcodeView(selectedUser) {
    this.root.html(passcodeTemplate(selectedUser));
    this.root.find("#Passcode").focus();
  }
};

api.listUsers((users) => $(() => LoginPage.init(users)));