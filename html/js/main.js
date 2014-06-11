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
			display(result);
		},
		"json"
	);
}

function display(result){
	document.body.innerHTML = "";
	var ul = document.createElement("ul");
	ul.id = "files_list";
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
			browse(path);
		}.bind(result.browse_command.path);
	}
	for(var i=0; i<result.browse_command.results.length; i++){
		var element = result.browse_command.results[i];
		var li = document.createElement("li");
		li.className = "files_item";
		var div = document.createElement("div");
		div.innerHTML = element.name;
		if (element.isDir){
			li.onclick = function(path){
				if ("/" != path.charAt(path.length - 1)){
					path = path + "/";
				}
				browse(path + this.name);
			}.bind(element, result.browse_command.path);
		}
		li.appendChild(div);
		ul.appendChild(li);
	}
	document.body.appendChild(ul);
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