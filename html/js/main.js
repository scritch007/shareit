var current_folder = null;

function browse(path){
	command = {
		name: "browser.browse",
		browse_command:{
			"path": path
		}
	}
	$.post(
		"/commands",
		JSON.stringify(command),
		function(result){
			current_folder = path;
			display(result);
		},
		"json"
	);
}

function createNavElement(name){
	var element = document.createElement("li");
	var a = document.createElement("a")
	if (null == name || undefined == name){
		var i = document.createElement("i");
		i.className = "icon-home";
		i.style.fontSize = "21px";
		a.appendChild(i);
	}else{
		a.innerHTML = name
	}
	element.appendChild(a);
	return element;
}
function display(result){
	var path = result.browse_command.path;
	if ("/" != path.charAt(path.length - 1)){
		path = path + "/";
	}
	var nav = document.getElementById("navigation_bar");
	var splitted_path = path.split("/");
	var nav_elements = splitted_path.slice(1, splitted_path.length -1);
	nav.innerHTML = "";
	var nav_ul = document.createElement("ul");
	nav.appendChild(nav_ul);
	var current_nav_element = createNavElement();
	current_nav_element.onclick = function(path){
		browse(path);
	}.bind(current_nav_element, "/");
	nav_ul.appendChild(current_nav_element);
	for (var i=0; i<nav_elements.length; i++){
		current_nav_element = createNavElement(nav_elements[i]);
		current_nav_element.onclick = function(path){
			browse(path);
		}.bind(current_nav_element, "/" + nav_elements.slice(0, i+1).join("/"));
		nav_ul.appendChild(current_nav_element);
	}
	if (null != current_nav_element){
		current_nav_element.className += " active";
	}
	var browse_div = document.getElementById("browsing");
	browse_div.className = "listview";
	browsing.innerHTML = "";

	if ("/" != result.browse_command.path){
		var a = document.createElement("a")
		a.className = "list";
		var img = document.createElement("img");
		img.src = "img/Folder.svg";
		img.className = "icon";
		var parent = document.createElement("div");
		parent.className = "files_item list-content bg-lightBlue";
		parent.appendChild(img);
		var parentLabel = document.createElement("span");

		parentLabel.innerHTML = "..";
		parentLabel.className = "list-title";
		parent.onclick = function(){
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
		}.bind(result.browse_command.path);
		parent.appendChild(parentLabel);
		a.appendChild(parent);
		browse_div.appendChild(a);
	}
	for(var i=0; i<result.browse_command.results.length; i++){
		var element = result.browse_command.results[i];
		element_path = path + element.name;
		var a = document.createElement("a");
		a.className = "list shadow";
		var img = document.createElement("img");
		img.className = "icon";
		var li = document.createElement("div");
		li.className = "files_item list-content bg-lightBlue";
		li.appendChild(img);
		var elementDiv = document.createElement("div");
		elementDiv.className = "data";
		var elementNameLabel = document.createElement("span");
		elementNameLabel.className = "list-title";
		elementNameLabel.innerHTML = element.name;
		elementDiv.appendChild(elementNameLabel);
		var buttonDiv = document.createElement("div");

		var deleteButton = document.createElement("input");
		deleteButton.type = "button";
		deleteButton.value = "delete";
		deleteButton.onclick = function(path, event){
			event.stopPropagation();
			document.body.appendChild(deletePopup(path));
		}.bind(element, element_path);
		buttonDiv.appendChild(deleteButton);
		elementDiv.appendChild(buttonDiv);

		if (element.isDir){
			li.onclick = function(path){
				browse(path);
			}.bind(element, element_path);
			img.src = "img/Folder.svg";
		}else{
			img.src = "img/Files.svg";
			var downloadButton = document.createElement("input");
			downloadButton.type = "button";
			downloadButton.value = "download";
			downloadButton.className = "button primary";
			downloadButton.onclick = function(path, event){
				event.stopPropagation()
				command = {
					name: "browser.download_link",
					download_link_command:{
						"path": path
					}
				}
				$.post(
					"/commands",
					JSON.stringify(command),
					function(result){
						console.log(result.download_link_command.download_link);
						document.body.appendChild(downloadPopup(path, result.download_link_command.download_link));
					},
					"json"
				);
			}.bind(element, element_path);
			buttonDiv.appendChild(document.createTextNode('\u00A0'));
			buttonDiv.appendChild(downloadButton);
		}
		li.appendChild(elementDiv);
		a.appendChild(li);
		browse_div.appendChild(a);
	}
	//browse_div.appendChild(ul);
}

function createFolder(){
	//Create a popup div to enter the name of the folder to create
	var createFolderPopup = document.createElement("div");
	createFolderPopup.className = "popup";
	var folderNameLabel = document.createElement("label");
	folderNameLabel.innerHTML = "Folder Name";
	var folderNameInput = document.createElement("input");
	folderNameInput.type = "text";
	var nameDiv = document.createElement("div");
	nameDiv.appendChild(folderNameLabel);
	nameDiv.appendChild(folderNameInput);
	createFolderPopup.appendChild(nameDiv);

	var buttonDiv = document.createElement("div");
	var cancelButton = document.createElement("input");
	cancelButton.type = "button";
	cancelButton.value = "Cancel";
	cancelButton.onclick = function(){
		createFolderPopup.parentNode.removeChild(createFolderPopup);
	}
	buttonDiv.appendChild(cancelButton);
	var goButton = document.createElement("input");
	goButton.type = "button";
	goButton.value = "Create";
	goButton.onclick = function(){
		goButton.disabled = true;
		cancelButton.disabled = true;
		path = current_folder;
		if ("/" != path.charAt(path.length - 1)){
			path = path + "/";
		}
		command = {
			name: "browser.create_folder",
			create_folder_command:{
				"path": path + folderNameInput.value
			}
		}
		$.post(
			"/commands",
			JSON.stringify(command),
			function(result){
				browse(current_folder);
				createFolderPopup.parentNode.removeChild(createFolderPopup);
			},
			"json"
		);
	}
	buttonDiv.appendChild(goButton);
	buttonDiv.className = "button_div";
	createFolderPopup.appendChild(buttonDiv);
	document.body.appendChild(createFolderPopup);
	folderNameInput.focus();
}

function downloadPopup(path, download_link){
	var window_div = document.createElement("div");
	window_div.className = "window shadow";
	window_div.id = "download_link_popup";
	var caption_div = document.createElement("div");
	caption_div.className = "caption";
	//Caption definition
	var caption_span = document.createElement("span");
	caption_span.className = "icon icon-windows";
	caption_div.appendChild(caption_span);
	var caption_title = document.createElement("div");
	caption_title.className = "title";
	caption_title.innerHTML = "Download " + path;
	caption_div.appendChild(caption_title);
	var caption_close_button = document.createElement("button");
	caption_close_button.type = "button";
	caption_close_button.className = "btn-close";
	caption_close_button.onclick = function(){
		window_div.parentNode.removeChild(window_div);
	}
	caption_div.appendChild(caption_close_button);
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

function deletePopup(path){
	var window_div = document.createElement("div");
	window_div.className = "window shadow";
	window_div.id = "delete_item_popup";
	var caption_div = document.createElement("div");
	caption_div.className = "caption";
	//Caption definition
	var caption_span = document.createElement("span");
	caption_span.className = "icon icon-windows";
	caption_div.appendChild(caption_span);
	var caption_title = document.createElement("div");
	caption_title.className = "title";
	caption_title.innerHTML = "Delete " + path;
	caption_div.appendChild(caption_title);
	var caption_close_button = document.createElement("button");
	caption_close_button.type = "button";
	caption_close_button.className = "btn-close";
	caption_div.appendChild(caption_close_button);
	caption_close_button.onclick = function(){
		window_div.parentNode.removeChild(window_div);
	}
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
		command = {
			name: "browser.delete_item",
			delete_command:{
				"path": path
			}
		}
		$.post(
			"/commands",
			JSON.stringify(command),
			function(result){
				window_div.parentNode.removeChild(window_div);
				browse(current_folder);
			},
			"json"
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
