var current_folder = null;

var displayTheme = null;

var mainWindow = null;

function setPopup(popup){
	mainWindow.innerHTML = "";
	mainWindow.appendChild(popup);
}

function init(){
	displayTheme = new WualaDisplay();
	mainWindow = document.getElementById("window_popup_id");
	$.getJSON("/auths", function(result){
		HandleAuthsResult(result);
	});
	browse("/");
}
function browse(path){
	sendRequest("/commands", "POST",
		{
			name: "browser.browse",
			browse_command:{
				"path": path
			}
		},
		function(result){
			current_folder = path;
			display(result);
		}
	)
}

function display(result){
	var path = result.browse_command.path;
	if ("/" != path.charAt(path.length - 1)){
		path = path + "/";
	}
	displayTheme.GetBrowsingPathElement(path, browse);
	var elementListObject = displayTheme.GetFilesListElement(path);

	if ("/" != result.browse_command.path){
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
			}.bind(result.browse_command.path)
		);
	}
	for(var i=0; i<result.browse_command.results.length; i++){
		var element = result.browse_command.results[i];
		element_path = path + element.name;
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
				sendRequest("/commands",
					"POST",
					{
						name: "browser.download_link",
						download_link_command:{
							"path": path
						}
					},
					function(result){
						console.log(result.download_link_command.download_link);
						setPopup(downloadPopup(path, result.download_link_command.download_link.link));
					}
				);
			}.bind(element, element_path);
		}
		displayTheme.AddElement(elementListObject, element, element.name, browseCB, downloadCB, deleteCB);
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
		sendRequest(
			"/commands",
			"POST",
			{
				name: "browser.create_folder",
				create_folder_command:{
					"path": path + folderNameInput.value
				}
			},
			function(result){
				browse(current_folder);
				createFolderPopup.parentNode.removeChild(createFolderPopup);
			}
		)
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
	url_div.appendChild(document.createTextNode('\u00A0'));
	var download_link_download_button = document.createElement("span");
	//download_link_download_button.type = "button";
	download_link_download_button.className = "icon icon-download";
	download_link_download_button.onclick = function(){
		window.open(dlink);
	}
	url_div.appendChild(download_link_download_button);
	display_content_div.appendChild(url_div);
	content_div.appendChild(display_content_div);
	window_div.appendChild(content_div);
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
	ok_button.className = "button primary";
	ok_button.onclick = function(){
		sendRequest(
			"/commands",
			"POST",
			{
				name: "browser.delete_item",
				delete_command:{
					"path": path
				}
			},
			function(result){
				window_div.parentNode.removeChild(window_div);
				browse(current_folder);
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
	cancel_button.className = "button";
	cancel_button.onclick = function(){
		window_div.parentNode.removeChild(window_div);
	}
	buttonDiv.appendChild(cancel_button);
	content_div.appendChild(buttonDiv);
	window_div.appendChild(content_div);
	return window_div;
}
/*
$.ajax({
    beforeSend: function(xhr) {
        xhr.setRequestHeader('X-HTTP-Method-Override', 'PUT');
    },
    type: 'POST',
    url: '/someurl',
    success: function(data){
        // do something...
    }
});
*/
