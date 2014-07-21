var availableAuths = [];

function HandleAuthsResult(result){
	console.log(result);
	for (var i =0; i < result.length; i++){
		switch(result[i]){
			case "DummyAuth":
			availableAuths.push(result[i]);
			break;
			default:
			console.log("Unknown authentication type " + result[i]);
			break;
		}
	}
}

function checkAuth(callback){
	if (undefined != sessionStorage.Authentication){
		//Hide login button and show logout one
		document.getElementById("logout").style.display="";
		document.getElementById("login").style.display="none";
		authorizationToken = sessionStorage.Authentication;
		//Todo ask for who you really are
		callback(sessionStorage.Authentication);
	}else{
		callback(null);
	}
}

function validatePassword(){
	var pass2=document.getElementById("signupPasswordCheck").value;
	var pass1=document.getElementById("signupPassword").value;
	if(pass1!=pass2)
	    document.getElementById("signupPasswordCheck").setCustomValidity("Passwords Don't Match");
	else
	    document.getElementById("signupPasswordCheck").setCustomValidity('');
	//empty string means no validation error
}

function signup(){
	var signupWindow = document.createElement("form");
	signupWindow.id="signupWindow";
	signupWindow.className = "window shadow";
	var caption = Caption("Sign up");
	signupWindow.appendChild(caption);
	var content_div = document.createElement("div");
	content_div.className = "content";
	signupWindow.appendChild(content_div);
	//Do not send the request the onclick will do it
	signupWindow.onsubmit = function(){return false;};
	//Content
	var loginDiv = document.createElement("div");
	var loginLabel = document.createElement("label");
	loginLabel.innerHTML = "Login";
	loginLabel.htmlFor="loginInput";
	var loginInput = document.createElement("input");
	signupWindow.focusElement = loginInput;
	loginInput.id="loginInput";
	loginInput.type = "text";
	loginInput.name="fname";
	loginInput.placeholder = "Enter your login";
	loginInput.setAttribute("required", true);
	loginDiv.appendChild(loginLabel);
	loginDiv.appendChild(loginInput);
	content_div.appendChild(loginDiv);
	var passwordDiv = document.createElement("div");
	var passwordLabel = document.createElement("label");
	passwordLabel.innerHTML = "Password";
	var passwordInput = document.createElement("input");
	passwordInput.type = "password";
	passwordInput.id = "signupPassword";
	passwordInput.placeholder = "Enter your password";
	passwordInput.setAttribute("required", true);
	passwordDiv.appendChild(passwordLabel);
	passwordDiv.appendChild(passwordInput);
	content_div.appendChild(passwordDiv);

	var passwordcheckDiv = document.createElement("div");
	var passwordcheckLabel = document.createElement("label");
	passwordcheckLabel.innerHTML = "Password Verification";
	var passwordcheckInput = document.createElement("input");
	passwordcheckInput.type = "password";
	passwordcheckInput.id = "signupPasswordCheck";
	passwordcheckInput.placeholder = "Enter same password";
	passwordcheckInput.setAttribute("required", true);
	passwordcheckDiv.appendChild(passwordcheckLabel);
	passwordcheckDiv.appendChild(passwordcheckInput);
	content_div.appendChild(passwordcheckDiv);

	passwordInput.onchange = validatePassword;
    passwordcheckInput.onchange = validatePassword;

	var emailDiv = document.createElement("div");
	var emailLabel = document.createElement("label");
	emailLabel.innerHTML = "Email";
	emailLabel.htmlFor = "signupEmail"
	var emailInput = document.createElement("input");
	emailInput.type = "email";
	emailInput.name = "email";
	emailInput.id = "signupEmail";
	emailInput.placeholder = "Enter your email";
	emailInput.setAttribute("required", true);
	emailDiv.appendChild(emailLabel);
	emailDiv.appendChild(emailInput);
	content_div.appendChild(emailDiv);

	var buttonDiv = document.createElement("div");
	buttonDiv.className = "footer";
	var cancelButton = document.createElement("a");
	cancelButton.type = "button small";
	cancelButton.innerHTML = "Cancel";
	cancelButton.className = "button small";
	cancelButton.onclick = function(){
		signupWindow.parentNode.removeChild(signupWindow);
	}
	buttonDiv.appendChild(cancelButton);
	var goButton = document.createElement("input");
	goButton.type = "submit";
	goButton.value = "Create";
	goButton.onclick = function(){
		if(!signupWindow.checkValidity())
		{
			return;
		}
		goButton.style.disabled = true;
		cancelButton.style.disabled = true;
		sendRequest({
				url:"auths/DummyAuth/create",
				method:"POST",
				data: {
					"login": loginInput.value,
					"password": passwordInput.value,
					"email": emailInput.value
				},
				onSuccess: function(result){
					console.log("Account created");
					signupWindow.parentNode.removeChild(signupWindow);
				}
			}
		);
	}
	goButton.className = "button small";
	buttonDiv.appendChild(goButton);
	signupWindow.appendChild(buttonDiv);
	setPopup(signupWindow);
}

function login(){
	var loginWindow = document.createElement("form");

	loginWindow.className = "window shadow";
	loginWindow.onsubmit = function(){return false;};
	var caption = Caption("Log in");
	loginWindow.appendChild(caption);
	var contentDiv = document.createElement("div");
	loginWindow.appendChild(contentDiv);
	var loginDiv = document.createElement("div");
	var loginLabel = document.createElement("label");
	loginLabel.htmlFor="loginInput";
	loginLabel.innerHTML = "Login";
	var loginInput = document.createElement("input");
	loginInput.type = "text";
	loginInput.setAttribute("required", true);
	loginWindow.focusElement = loginInput;
	loginInput.placeholder = "Enter your login";
	loginDiv.appendChild(loginLabel);
	loginDiv.appendChild(loginInput);
	contentDiv.appendChild(loginDiv);
	var passwordDiv = document.createElement("div");
	var passwordLabel = document.createElement("label");
	passwordLabel.htmlFor="passwordInput";
	passwordLabel.innerHTML = "Password";
	var passwordInput = document.createElement("input");
	passwordInput.type = "password";
	passwordInput.setAttribute("required", true);
	passwordInput.placeholder = "Enter your login";
	passwordDiv.appendChild(passwordLabel);
	passwordDiv.appendChild(passwordInput);
	contentDiv.appendChild(passwordDiv);
	var buttonDiv = document.createElement("div");
	buttonDiv.className = "footer";
	var cancelButton = document.createElement("a");
	cancelButton.type = "button small";
	cancelButton.innerHTML = "Cancel";
	cancelButton.className = "button small";
	cancelButton.onclick = function(){
		loginWindow.parentNode.removeChild(loginWindow);
	}
	buttonDiv.appendChild(cancelButton);
	var goButton = document.createElement("input");
	goButton.type = "submit";
	goButton.value = "Login";
	goButton.onclick = function(){
		if(!loginWindow.checkValidity())
		{
			return;
		}
		goButton.disabled = true;
		cancelButton.disabled = true;
		//Get the challenge
		sendRequest(
			{
				url:"auths/DummyAuth/get_challenge",
				method:"GET",
				onSuccess: function(result){
					//TODO at one point we should hash the challenge but never mind for now :)
					sendRequest(
						{
							url: "auths/DummyAuth/auth",
							method: "POST",
							data: {
								"login": loginInput.value,
								"challenge_hash": result.challenge + ":" + passwordInput.value,
								"ref": result.ref
							},
							onSuccess: function(result){
								//Hide login button and show logout one
								document.getElementById("logout").style.display="";
								document.getElementById("login").style.display="none";
								//Set the Global Header
								authorizationToken = result.authentication_header;
								loginWindow.parentNode.removeChild(loginWindow);
								sessionStorage.Authentication = authorizationToken;
								browse(current_folder);
							}
						}
					);
				}
			}
		);
	}
	goButton.className = "button small";
	buttonDiv.appendChild(goButton);
	loginWindow.appendChild(buttonDiv);
	return setPopup(loginWindow);
}

function logout(){
	delete sessionStorage.Authentication;
	authorizationToken = null;
	//Hide login button and show logout one
	document.getElementById("logout").style.display="none";
	document.getElementById("login").style.display="";
	browse(current_folder);
}
