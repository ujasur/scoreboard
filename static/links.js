toastr.options.closeDuration = 200;
toastr.options.timeOut = 4000;
api.defaultErrorHandler = res => toastr.error(res.error.error);

var LinksPage = {
  tpl: {
    Links: Handlebars.compile(`
      <h6 class="mt-3">Links <small class="float-right">avail. {{links.length}} links</small></h6>
      <ul class="list-group">
        {{#each links}}
        <li class="list-group-item clearfix">
          <a href="{{this.uri}}" class="list-item-txt float-left text-truncate" style="max-width:500px" title="{{this.uri}}">{{this.display_name}}</a>
          <span class="float-right">
            <button class="del-link-btn btn btn-sm btn-outline-danger" data-id="{{this.id}}">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16" fill="#fff"><path fill-rule="evenodd" d="M3.72 3.72a.75.75 0 011.06 0L8 6.94l3.22-3.22a.75.75 0 111.06 1.06L9.06 8l3.22 3.22a.75.75 0 11-1.06 1.06L8 9.06l-3.22 3.22a.75.75 0 01-1.06-1.06L6.94 8 3.72 4.78a.75.75 0 010-1.06z"></path></svg>
            </button>
          </span>
        </li>  
        {{else}}
        <li class="list-group-item clearfix">
          <span class="list-item-txt font-italic text-center">There are no links available</span>
        </li>
        {{/each}}
      </ul>
    `)
  },
  init(links) {
    this.links = links || [];
    this.createBtn = $('#CreateLinkBtn');
    this.inputName = $('#InputName');
    this.inputUri = $('#InputUri');
    this.linksContainer = $("#LinksContainer");

    this.createBtn.on('click', () => {
      var newLink = {
        uri: this.inputUri.val() || '',
        display_name: this.inputName.val() || '',
      };
      if (!newLink.display_name.length) {
        this.inputName.addClass('is-invalid');
        return;
      }
      if (!newLink.uri.length) {
        this.inputUri.addClass('is-invalid');
        return;
      }

      this.createBtn.attr('disabled', true);
      api.linkAdd(newLink, (link) => {
        this.createBtn.removeAttr('disabled');
        this.inputName.val('');
        this.inputUri.val('');
        this.links.push(link);
        this.render();
      }, () => {
          this.createBtn.removeAttr('disabled');
          return true;
        });
    });

    this.linksContainer.on('click', '.del-link-btn', (event) => {
      var target = $(event.target).closest('.del-link-btn');
      var id = target.data('id');
      api.linkRemove(id, () => {
        this.links = this.links.filter(l => l.id !== id);
        this.render();
      });
    });

    this.render();
  },

  render() {
    this.linksContainer.html(this.tpl.Links({ links: this.links }));
  }
}
api.linkList((links) => LinksPage.init(links));