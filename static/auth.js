function logout() {
  window.localStorage.removeItem('user');
  window.location = '/ui/login';
}

var userData = window.localStorage.getItem('user');
if (!userData) {
  logout();
}

var user = JSON.parse(userData);  
$("#LogoutBtn").text('Logout ' + user.name).removeClass('d-none');
$.ajaxSetup({ headers: { 'authorization': user.token } });