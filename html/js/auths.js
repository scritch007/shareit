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
	if (undefined != localStorage.Authentication){
		//Hide login button and show logout one
		document.getElementById("logout").style.display="";
		document.getElementById("login").style.display="none";
		authorizationToken = localStorage.Authentication;
		//Todo ask for who you really are
		callback(localStorage.Authentication);
	}else{
		callback(null);
	}
}

function signup(){
	var signupWindow = document.createElement("div");
	signupWindow.className = "window shadow";
	var caption = Caption("Sign up");
	signupWindow.appendChild(caption);
	var content_div = document.createElement("div");
	content_div.className = "content";
	signupWindow.appendChild(content_div);
	//Content
	var loginDiv = document.createElement("div");
	var loginLabel = document.createElement("label");
	loginLabel.innerHTML = "Login";
	var loginInput = document.createElement("input");
	loginInput.type = "text";
	loginDiv.appendChild(loginLabel);
	loginDiv.appendChild(loginInput);
	content_div.appendChild(loginDiv);
	var passwordDiv = document.createElement("div");
	var passwordLabel = document.createElement("label");
	passwordLabel.innerHTML = "Password";
	var passwordInput = document.createElement("input");
	passwordInput.type = "text";
	passwordDiv.appendChild(passwordLabel);
	passwordDiv.appendChild(passwordInput);
	content_div.appendChild(passwordDiv);

	var passwordcheckDiv = document.createElement("div");
	var passwordcheckLabel = document.createElement("label");
	passwordcheckLabel.innerHTML = "Password Verification";
	var passwordcheckInput = document.createElement("input");
	passwordcheckInput.type = "text";
	passwordcheckDiv.appendChild(passwordcheckLabel);
	passwordcheckDiv.appendChild(passwordcheckInput);
	content_div.appendChild(passwordcheckDiv);


	var emailDiv = document.createElement("div");
	var emailLabel = document.createElement("label");
	emailLabel.innerHTML = "Email";
	var emailInput = document.createElement("input");
	emailInput.type = "text";
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
	var goButton = document.createElement("a");
	goButton.innerHTML = "Create";
	goButton.onclick = function(){
		goButton.disabled = true;
		cancelButton.disabled = true;
		$.post("/auths/DummyAuth/create", JSON.stringify({
			"login": loginInput.value,
			"password": passwordInput.value,
			"email": emailInput.value
			}),
			function(result){
				console.log("Account created");
			}
		);
	}
	goButton.className = "button small";
	buttonDiv.appendChild(goButton);
	signupWindow.appendChild(buttonDiv);
	setPopup(signupWindow);
}

function login(){
	var loginWindow = document.createElement("div");
	loginWindow.className = "window shadow";
	var caption = Caption("Log in");
	loginWindow.appendChild(caption);
	var contentDiv = document.createElement("div");
	loginWindow.appendChild(contentDiv);
	var loginDiv = document.createElement("div");
	var loginLabel = document.createElement("label");
	loginLabel.innerHTML = "Login";
	var loginInput = document.createElement("input");
	loginInput.type = "text";
	loginDiv.appendChild(loginLabel);
	loginDiv.appendChild(loginInput);
	contentDiv.appendChild(loginDiv);
	var passwordDiv = document.createElement("div");
	var passwordLabel = document.createElement("label");
	passwordLabel.innerHTML = "Password";
	var passwordInput = document.createElement("input");
	passwordInput.type = "text";
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
	var goButton = document.createElement("a");
	goButton.innerHTML = "Create";
	goButton.onclick = function(){
		goButton.disabled = true;
		cancelButton.disabled = true;
		//Get the challenge
		sendRequest("/auths/DummyAuth/get_challenge","GET", null, function(result){
			//TODO at one point we should hash the challenge but never mind for now :)
			sendRequest("/auths/DummyAuth/auth", "POST", {
				"login": loginInput.value,
				"challenge_hash": result.challenge + ":" + passwordInput.value,
				"ref": result.ref
			}, function(result){
				//Hide login button and show logout one
				document.getElementById("logout").style.display="";
				document.getElementById("login").style.display="none";
				//Set the Global Header
				authorizationToken = result.authentication_header;
				loginWindow.parentNode.removeChild(loginWindow);
				localStorage.Authentication = authorizationToken;
			});
		});
	}
	goButton.className = "button small";
	buttonDiv.appendChild(goButton);
	loginWindow.appendChild(buttonDiv);
	return setPopup(loginWindow);
}

function logout(){
	delete localStorage.Authentication;
	//Hide login button and show logout one
	document.getElementById("logout").style.display="none";
	document.getElementById("login").style.display="";
}

var authorizationToken = null;

function sendRequest(obj){
	request = {
        type:obj.method,
        beforeSend: function (request)
        {
        	if (null != authorizationToken){
        		request.setRequestHeader("Authentication", authorizationToken);
        	}

        },
        url: obj.url,
        processData: false,
        success: function(results) {
			obj.onSuccess(results);
        },
        error: function(request, status, error){
        	if ((null != errorCB) && (undefined != errorCB)){
        		obj.errorCB(status, error);
        	}
        },
        dataType:"json"
    }
    if (null != obj.data || undefined != obj.data){
    	request.data = JSON.stringify(obj.data);
    }
   	$.ajax(request);
}
