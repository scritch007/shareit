var current_folder = null;

var displayTheme = null;

var mainWindow = null;

var queryString = {};

function getQueryString() {
  var result = {}, queryString = location.search.slice(1),
      re = /([^&=]+)=([^&]*)/g, m;

  while (m = re.exec(queryString)) {
    result[decodeURIComponent(m[1])] = decodeURIComponent(m[2]);
  }
  return result;
}

function setPopup(popup){
	mainWindow.innerHTML = "";
	mainWindow.appendChild(popup);
	if (undefined != popup.focusElement){
		popup.focusElement.focus();
	}
}

function init(){
	queryString = getQueryString();
	displayTheme = new WualaDisplay();
	mainWindow = document.getElementById("window_popup_id");
	$.getJSON("auths", function(result){
		HandleAuthsResult(result);
		checkAuth(function(loggedUser){
			if (null != loggedUser){
				//Display the current user name somewhere
			}
			browse("/");
		});
	});
}
function browse(path){
	var command = {
		name: "browser.list",
		browser:{
			list:{
				"path": path
			}
		}
	};
	request = {
		data: command,
		onSuccess: function(result){
			current_folder = path;

			display(result, undefined==result.auth_key);
		},
		onError: function(request, status, error){
			if (401 == request.status){
				logout();
				browse(path);
			}
		},
		poll:true
	}
	sendCommand(request);
}

function display(result, add_share_callback){
	var path = result.browser.list.path;
	if ("/" != path.charAt(path.length - 1)){
		path = path + "/";
	}
	displayTheme.GetBrowsingPathElement(path, browse);
	var elementListObject = displayTheme.GetFilesListElement(path);

	if ("/" != path){
		displayTheme.AddElement(elementListObject, null, "..",
			function(event){
				var path = this;
				if ("/" == path.charAt(path.length - 1)){
					path = path.substr(0, path.length - 1);
				}
				var split = path.split("/");
				path = split.slice(0, split.length - 1).join("/");
				if (path == "")
				{
					path = "/";
				}
				browse(path);
			}.bind(path)
		);
	}
	for(var i=0; i<result.browser.list.results.length; i++){
		var element = result.browser.list.results[i];
		element_path = path + element.name;
		element.element_path = element_path;
		var downloadCB = null;
		var browseCB = null;
		var deleteCB = function(path, event){
			event.stopPropagation();
			setPopup(deletePopup(path));
		}.bind(element, element_path);

		if (element.isDir){
			browseCB = function(path, event){
				browse(path);
			}.bind(element, element_path);
		}else{
			downloadCB = function(path, event){
				event.stopPropagation()
				sendCommand(
					{
						data: {
							name: "browser.download_link",
							browser:{
								download_link:{
									"path": path
								}
							}
						},
						onSuccess: function(result){
							console.log(result.browser.download_link.download_link);
							setPopup(downloadPopup(path, result.browser.download_link.download_link.link));
						}
					}
				);
			}.bind(element, element_path);
		}
		var shareCB = null;
		if (null != authorizationToken && add_share_callback){
			shareCB = function(){
				//Check if we've got a sharelink save
				sendCommand({
					data : {
						name: "share_link.get",
						share_link:{
							get : {
								path: this.element_path
							}
						}
					},
					onSuccess: function(result){
						setPopup(sharePopup(this, result));
					}.bind(this),
					onError: function(result){
						setPopup(sharePopup(this, null));
					}.bind(this)
				});
			}.bind(element);
		}
		displayTheme.AddElement(elementListObject, element, element.name, browseCB, downloadCB, deleteCB, shareCB);
	}
	if (0 == result.browser.list.results.length){
		//Display that this is empty
		displayTheme.AddEmptyList(elementListObject);
	}
	//browse_div.appendChild(ul);
	var mainDisplay = document.getElementById("browsing");
	mainDisplay.innerHTML = "";
	mainDisplay.appendChild(displayTheme.GetListDisplayComponent(elementListObject));
}

function createFolder(){
	//Create a popup div to enter the name of the folder to create
	var createFolderPopup = document.createElement("div");
	createFolderPopup.className = "window shadow";


	var caption_div = Caption("Create New Folder");
	createFolderPopup.appendChild(caption_div);
	var folderNameLabel = document.createElement("label");
	folderNameLabel.innerHTML = "Folder Name";
	var folderNameInput = document.createElement("input");
	folderNameInput.type = "text";
	var nameDiv = document.createElement("div");
	nameDiv.appendChild(folderNameLabel);
	nameDiv.appendChild(folderNameInput);
	createFolderPopup.appendChild(nameDiv);

	var buttonDiv = document.createElement("div");
	buttonDiv.className = "footer";
	var cancelButton = document.createElement("a");
	cancelButton.type = "button small";
	cancelButton.innerHTML = "Cancel";
	cancelButton.className = "button small";
	cancelButton.onclick = function(){
		createFolderPopup.parentNode.removeChild(createFolderPopup);
	}
	buttonDiv.appendChild(cancelButton);
	var goButton = document.createElement("a");
	goButton.innerHTML = "Create";
	goButton.onclick = function(){
		goButton.disabled = true;
		cancelButton.disabled = true;
		path = current_folder;
		if ("/" != path.charAt(path.length - 1)){
			path = path + "/";
		}
		sendCommand(
			{
				data: {
					name: "browser.create_folder",
					browser:{
						create_folder:{
							"path": path + folderNameInput.value
						}
					}
				},
				onSuccess: function(result){
					browse(current_folder);
					createFolderPopup.parentNode.removeChild(createFolderPopup);
				}
			}
		);
	}
	goButton.className = "button small";
	buttonDiv.appendChild(goButton);
	createFolderPopup.appendChild(buttonDiv);
	setPopup(createFolderPopup);
	folderNameInput.focus();
}

function downloadPopup(path, download_link){
	var window_div = document.createElement("div");
	window_div.className = "window shadow";
	window_div.id = "download_link_popup";
	var caption_div = Caption("Download " + path)
	//End of Caption defintion
	window_div.appendChild(caption_div);
	var content_div = document.createElement("div");
	content_div.className = "content";
	content_div.id = "download_link_content";
	var display_content_div = document.createElement("div");
	var url_div = document.createElement("div");
	url_div.className = "input-control text";
	var download_link_input = document.createElement("input");
	download_link_input.type = "text";
	var dlink = location.protocol + "//" + location.host + "/downloads/" + download_link;
	download_link_input.value = dlink;
	url_div.appendChild(download_link_input);

	var downloadButton = document.createElement("a");
	downloadButton.className = "button small w-download";
	var i = document.createElement("i");
	i.className = "icon-arrow-down";
	downloadButton.appendChild(i);
	downloadButton.appendChild(i);
	url_div.appendChild(downloadButton);
	downloadButton.onclick = function(){
		window.open(dlink);
	}
	display_content_div.appendChild(url_div);
	content_div.appendChild(display_content_div);
	window_div.appendChild(content_div);
	return window_div;
}

function sharePopup(element, result){
	var share_link = null;
	if (null != result){
		share_link = result.share_link.get.result;
	}
	var window_div = document.createElement("div");
	window_div.className = "window shadow";
	window_div.id = "share_link_popup";
	var caption_div = Caption("Share " + element.name);
	//End of Caption defintion
	window_div.appendChild(caption_div);
	var content_div = document.createElement("div");
	content_div.className = "content";
	content_div.id = "share_link_content";

	var keyDiv = document.createElement("div");
	var keyLabel = document.createElement("div");
	keyLabel.innerHTML = "ShareLinkKey";
	keyDiv.appendChild(keyLabel);
	var keyInput = document.createElement("input");
	keyInput.id = "sharelinkkey";
	keyInput.type = "text";
	keyDiv.appendChild(keyInput);
	if (null != share_link){
		keyInput.value = share_link.key;
	}
	content_div.appendChild(keyDiv);

	var shareLinkTypeDiv = document.createElement("div");
	var shareLinkTypeLabel = document.createElement("label");
	shareLinkTypeLabel.innerHTML = "Share Link Type";
	shareLinkTypeDiv.appendChild(shareLinkTypeLabel);
	var shareLinkTypeSelect = document.createElement("select");
	var shareLinkType = [ "key", "authenticated", "restricted"];
	for (var i=0; i<shareLinkType.length; i++){
		var option = document.createElement("option");
		option.value = shareLinkType[i];
		option.innerHTML = shareLinkType[i];
		shareLinkTypeSelect.appendChild(option);
	}
	shareLinkTypeSelect.onchange = function(event){
		for (var i=0; i < shareLinkTypeSelect.options.length; i++){
			var option = shareLinkTypeSelect.options[i];
			if (option.selected){
				document.getElementById("share_link_" + option.value).style.display="";
			}else{
				document.getElementById("share_link_" + option.value).style.display="none";
			}
		}
	}
	shareLinkTypeDiv.appendChild(shareLinkTypeSelect);
	content_div.appendChild(shareLinkTypeDiv);
	var shareLinkSpecificDiv = document.createElement("div");
	var shareLinkDivKey = document.createElement("div");
	shareLinkDivKey.id = "share_link_key";
	shareLinkSpecificDiv.appendChild(shareLinkDivKey);
	var shareLinkDivAuthenticated = document.createElement("div");
	shareLinkDivAuthenticated.style.display="none";
	shareLinkDivAuthenticated.id = "share_link_authenticated";
	shareLinkSpecificDiv.appendChild(shareLinkDivAuthenticated);

	var shareLinkDivRestricted = document.createElement("div");
	shareLinkDivRestricted.style.display="none";
	shareLinkSpecificDiv.appendChild(shareLinkDivRestricted);
	shareLinkDivRestricted.id = "share_link_restricted";

	//Add the list of users added to this share link
	var usersUl = document.createElement("ul");
	shareLinkDivRestricted.appendChild(usersUl);
	if (null != share_link && "restricted" == share_link.type){
		for(var i=0; i < share_link.user_list.length; i++){
			var userLi = document.createElement("li");
			userLi.innerHTML = share_link.user_list[i];
			userUl.appendChild(userLi);
		}
	}

	//Share Restricted requires listing the users..
	var searchTimer = null;
	var searchUsersDiv = document.createElement("div");
	var searchUsersSpan = document.createElement("span");
	searchUsersDiv.appendChild(searchUsersSpan);
	var searchUsersInput = document.createElement("input");
	searchUsersSpan.appendChild(searchUsersInput);
	searchUsersInput.type = "list";
	searchUsersInput.setAttribute("list", "searchUserResults");
	var buttonPlus = document.createElement("button");
	buttonPlus.className = "button small w-expand-info";
	var iButtonPlus = document.createElement("i");
	iButtonPlus.className = "icon-plus";
	buttonPlus.appendChild(iButtonPlus);
	buttonPlus.onclick = function(event){
		//TODO add user to user list
		event.stopPropagation();
	}
	searchUsersSpan.appendChild(buttonPlus);

	var searchUsersResponse = document.createElement("datalist");
	searchUsersResponse.id = "searchUserResults";
	searchUsersSpan.appendChild(searchUsersResponse);
	searchUsersInput.onkeyup = function(){
		clearTimeout(searchTimer);
		searchTimer = setTimeout(
			function(){
				sendRequest(
					{
						url:"auths/list_users?search=" + searchUsersInput.value,
						method:"GET",
						onSuccess: function(result){
							searchUsersResponse.innerHTML = "";
							for(var i=0; i < result.length; i++){
								var label = document.createElement("option");
								label.value = result[i].name + "(" + result[i].id +")";
								searchUsersResponse.appendChild(label);
							}
						}
					}
				);
			},
			300);
	}
	shareLinkDivRestricted.appendChild(searchUsersDiv);
	var selectedUsers = document.createElement("div");
	shareLinkDivRestricted.appendChild(selectedUsers);

	content_div.appendChild(shareLinkSpecificDiv);
	window_div.appendChild(content_div);
	var buttonDiv = document.createElement("div");
	buttonDiv.className = "footer";

	var ok_button = document.createElement("input");
	ok_button.type = "button";
	ok_button.value = "Yes";
	ok_button.className = "button primary small";
	ok_button.onclick = function(){
		var cmd_name = null == share_link ? "create":"update";
		command = {
			name: "share_link." + cmd_name,
			share_link: {}
		};
		command.share_link[cmd_name] = {
			share_link: {
				path: current_folder +"/" + element.name,
				type: shareLinkTypeSelect.selectedOptions[0].value
			}
		};
		if ("restricted" == shareLinkTypeSelect.selectedOptions[0].value){
			//Add the users that have access to this share link

		}
		sendCommand(
			{
				data: command,
				poll: true,
				onSuccess:function(result){
					console.log(result);
					if (0 == result.state.status){
						keyInput.value = result.share_link.create.share_link.key;
					}
					//Else notify of an error...
				}
			}
		)
	};
	buttonDiv.appendChild(ok_button);
	var spacer = document.createTextNode('\u00A0');
	spacer.className="spacer";
	buttonDiv.appendChild(spacer);
	var cancel_button = document.createElement("input");
	cancel_button.type = "button";
	cancel_button.value = "No";
	cancel_button.className = "button small";
	cancel_button.onclick = function(){
		window_div.parentNode.removeChild(window_div);
	}
	buttonDiv.appendChild(cancel_button);
	window_div.appendChild(buttonDiv);
	return window_div;
}

function Caption(text){
	var caption_div = document.createElement("div");
	caption_div.className = "caption";
	//Caption definition
	var caption_span = document.createElement("span");
	caption_span.className = "icon icon-windows";
	caption_div.appendChild(caption_span);
	var caption_title = document.createElement("div");
	caption_title.className = "title";
	caption_title.innerHTML = text;
	caption_div.appendChild(caption_title);
	var caption_close_button = document.createElement("a");
	caption_close_button.className = "button small";
	var i =document.createElement("i");
	i.className = "icon-remove";
	caption_close_button.appendChild(i);
	caption_div.appendChild(caption_close_button);
	caption_close_button.onclick = function(){
		caption_div.parentNode.parentNode.removeChild(caption_div.parentNode);
	}
	return caption_div;
}

function deletePopup(path){
	var window_div = document.createElement("div");
	window_div.className = "window shadow";
	window_div.id = "delete_item_popup";
	var caption_div = Caption("Delete " + path);
	//End of Caption defintion
	window_div.appendChild(caption_div);
	var content_div = document.createElement("div");
	content_div.className = "content";
	content_div.id = "delete_item_content";
	var h3 = document.createElement("h3");
	h3.innerHTML = "Do you want to remove " + path;

	content_div.appendChild(h3);
	var buttonDiv = document.createElement("div");
	buttonDiv.className = "form-actions"

	var ok_button = document.createElement("input");
	ok_button.type = "button";
	ok_button.value = "Yes";
	ok_button.className = "button primary small";
	ok_button.onclick = function(){
		sendCommand(
			{
				data: {
					name: "browser.delete_item",
					browser:{
						"delete":{
							"path": path
						}
					}
				},
				onSuccess: function(result){
					window_div.parentNode.removeChild(window_div);
					browse(current_folder);
				}
			}
		);
	};
	buttonDiv.appendChild(ok_button);
	var spacer = document.createTextNode('\u00A0');
	spacer.className="spacer";
	buttonDiv.appendChild(spacer);
	var cancel_button = document.createElement("input");
	cancel_button.type = "button";
	cancel_button.value = "No";
	cancel_button.className = "button small";
	cancel_button.onclick = function(){
		window_div.parentNode.removeChild(window_div);
	}
	buttonDiv.appendChild(cancel_button);
	content_div.appendChild(buttonDiv);
	window_div.appendChild(content_div);
	return window_div;
}
function uploadFile(){
	//Create a popup div to enter the name of the folder to create
	var uploadFilePopup = document.createElement("form");
	uploadFilePopup.className = "window shadow";
	uploadFilePopup.onsubmit = function(){return false;};
	var caption_div = Caption("uploadFile");
	uploadFilePopup.appendChild(caption_div);
	var folderNameLabel = document.createElement("label");
	folderNameLabel.innerHTML = "Folder Name";
	var fileNameInput = document.createElement("input");
	fileNameInput.type = "file";
	fileNameInput.id = "files";
	fileNameInput.name = "file";
	fileNameInput.setAttribute("required", true);
	var nameDiv = document.createElement("div");
	nameDiv.appendChild(folderNameLabel);
	nameDiv.appendChild(fileNameInput);
	uploadFilePopup.appendChild(nameDiv);

	var buttonDiv = document.createElement("div");
	buttonDiv.className = "footer";
	var cancelButton = document.createElement("a");
	cancelButton.type = "button small";
	cancelButton.innerHTML = "Cancel";
	cancelButton.className = "button small";
	cancelButton.onclick = function(){
		uploadFilePopup.parentNode.removeChild(uploadFilePopup);
	}
	buttonDiv.appendChild(cancelButton);
	var goButton = document.createElement("input");
	goButton.type = "submit";
	goButton.value = "Upload";
	goButton.innerHTML = "Create";
	goButton.onclick = function(){
		if(!uploadFilePopup.checkValidity())
		{
			return;
		}
		goButton.disabled = true;
		cancelButton.disabled = true;
		path = current_folder;
		if ("/" != path.charAt(path.length - 1)){
			path = path + "/";
		}
		sendCommand(
			{
				data: {
					name: "browser.upload_file",
					browser:{
						upload_file:{
							"path": path + fileNameInput.files[0].name,
							"size": fileNameInput.files[0].size
						}
					}
				},
				onSuccess: function(result){
					console.log(JSON.stringify(result));

					var notification = new Notification({progressBar:true, name:fileNameInput.files[0].name});
					function notificationUpdate(file, uploadedSize){
						notification.progressBar.value = uploadedSize/this.size * 100;
					}

					//Now start the real work
					uploader = new ChunkedUploader(fileNameInput.files[0], {url: "/commands/" + result.command_id, progressCB:notificationUpdate.bind(fileNameInput.files[0], notification)});
					uploader.start();
					document.getElementById("notifications").appendChild(notification);
					uploadFilePopup.parentNode.removeChild(uploadFilePopup);
				}
			}
		);
	}
	goButton.className = "button small";
	buttonDiv.appendChild(goButton);
	uploadFilePopup.appendChild(buttonDiv);
	setPopup(uploadFilePopup);
	fileNameInput.focus();
}
