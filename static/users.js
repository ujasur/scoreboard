
  toastr.options.closeDuration = 200;
  toastr.options.timeOut = 4000;

  api.defaultErrorHandler = res => toastr.error(res.error.error);
  var UsersPage = {
    tpl: {
      Voters: Handlebars.compile(`
          <h6 class="mb-3">Voters <small class="float-right">avail. {{users.length}} users</small></h6>
          <ul class="list-group">
            {{#each users}}
            <li class="list-group-item clearfix">
              <span class="username list-item-txt float-left">{{this.name}}</span>
              <span class="float-right">
                <button class="del-user-btn btn btn-sm btn-outline-danger" data-username="{{this.name}}">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16" fill="#fff"><path fill-rule="evenodd" d="M3.72 3.72a.75.75 0 011.06 0L8 6.94l3.22-3.22a.75.75 0 111.06 1.06L9.06 8l3.22 3.22a.75.75 0 11-1.06 1.06L8 9.06l-3.22 3.22a.75.75 0 01-1.06-1.06L6.94 8 3.72 4.78a.75.75 0 010-1.06z"></path></svg>
                </button>
              </span>
            </li>  
            {{/each}}
          </ul>
        `),
      Masters: Handlebars.compile(`
          <h6 class="mb-3 mt-3">Masters</h6>
          <ul class="list-group">
            {{#each users}}
            <li class="list-group-item clearfix">
              <span class="list-item-txt">{{this.name}}</span>
            </li>  
            {{/each}}
          </ul>
        `)
    },

    init(users) {
      this.users = users || [];
      this.addUserBtn = $('#NewUserBtn');
      this.usernameInput = $('#NewUserInput');
      this.votersContainer = $("#VotersContainer");
      this.mastersContainer = $("#MastersContainer");

      this.addUserBtn.on('click', () => {
        var newUser = {
          name: this.usernameInput.val() || '',
          role: 'voter'
        };
        if (!newUser.name.length) {
          return;
        }
        var match = this.users.find((u) => u.name == newUser.name);
        if (match) {
          toastr.error("User " + newUser.name + " exists already.");
          return;
        }
        this.addUserBtn.attr('disabled', true);
        api.userAdd(newUser, () => {
          this.addUserBtn.removeAttr('disabled');
          this.usernameInput.val('');

          this.users.push(newUser);
          this.render();
          }, () => {
            this.addUserBtn.removeAttr('disabled');
            return true;
          });
      });

      this.votersContainer.on('click', '.del-user-btn', (event) => {
        var target = $(event.target).closest('.del-user-btn');
        var name = target.data('username');
        api.userRemove(name, () => {
          this.users = this.users.filter(u => u.name !== name);
          this.render();
        });
      });

      this.render();
    },

    render() {
      this.votersContainer.html(this.tpl.Voters({ users: this.users.filter(u => u.role == 'voter') }));
      this.mastersContainer.html(this.tpl.Masters({ users: this.users.filter(u => u.role == 'scrum_master') }));
    }
  }
  api.listUsers((users) => UsersPage.init(users));