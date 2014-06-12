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

function display(result){
	var browse_div = document.getElementById("browsing");
	browsing.innerHTML = "";
	var ul = document.createElement("ul");
	ul.id = "files_list";
	var path = result.browse_command.path;
	if ("/" != path.charAt(path.length - 1)){
		path = path + "/";
	}
	if ("/" != result.browse_command.path){
		var parent = document.createElement("li");
		var li = document.createElement("div");
		li.className = "files_item";
		parent.innerHTML = "..";
		li.appendChild(parent);
		ul.appendChild(parent);
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
	}
	for(var i=0; i<result.browse_command.results.length; i++){
		var element = result.browse_command.results[i];
		element_path = path + element.name;
		var li = document.createElement("li");
		li.className = "files_item";
		var elementDiv = document.createElement("div");
		var elementNameLabel = document.createElement("label");
		elementNameLabel.innerHTML = element.name;
		elementDiv.appendChild(elementNameLabel);
		var deleteButton = document.createElement("input");
		deleteButton.type = "button";
		deleteButton.value = "delete";
		deleteButton.onclick = function(path, event){
			event.stopPropagation()
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
					browse(current_folder);
				},
				"json"
			);
		}.bind(element, element_path);
		elementDiv.appendChild(deleteButton);
		if (element.isDir){
			li.onclick = function(path){
				browse(path);
			}.bind(element, element_path);
		}
		li.appendChild(elementDiv);
		ul.appendChild(li);
	}
	browse_div.appendChild(ul);
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
