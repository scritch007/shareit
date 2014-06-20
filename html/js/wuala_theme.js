function WualaDisplay(){

}

WualaDisplay.prototype.createNavElement = function(name){
	var element = document.createElement("li");
	var a = document.createElement("a")
	if (null == name || undefined == name){
		var i = document.createElement("i");
		i.className = "icon-home";
		//i.style.fontSize = "21px";
		a.appendChild(i);
	}else{
		a.innerHTML = name
	}
	element.appendChild(a);
	return element;
}

WualaDisplay.prototype.GetBrowsingPathElement = function(path, onBrowseChangeCB){
	var nav_bar = document.getElementById("navigation_bar");
	nav_bar.innerHTML = "";
	nav_bar.className = "row hide-for-medium-down";
	var nav = document.createElement("div");
	nav_bar.appendChild(nav);
	nav.className = "large-11 columns";
	nav.id = "breadcrumbs";
	var splitted_path = path.split("/");
	var nav_elements = splitted_path.slice(1, splitted_path.length -1);
	nav.innerHTML = "";
	var nav_ul = document.createElement("ul");
	nav_ul.className = "breadcrumbs";
	nav.appendChild(nav_ul);
	var current_nav_element = this.createNavElement();
	current_nav_element.onclick = function(path){
		onBrowseChangeCB(path);
	}.bind(current_nav_element, "/");
	nav_ul.appendChild(current_nav_element);
	for (var i=0; i<nav_elements.length; i++){
		current_nav_element = this.createNavElement(nav_elements[i]);
		current_nav_element.onclick = function(path){
			onBrowseChangeCB(path);
		}.bind(current_nav_element, "/" + nav_elements.slice(0, i+1).join("/"));
		nav_ul.appendChild(current_nav_element);
	}
	if (null != current_nav_element){
		current_nav_element.className += " active";
	}
}

WualaDisplay.prototype.GetFilesListElement = function(path){
	var browse_div = document.createElement("table");
	browse_div.className = "footable twelve";
	browsing.innerHTML = "";
	this.thead = document.createElement("thead");

	var tr = document.createElement("tr");
	this.thead.appendChild(tr);

	//For the mobile
	var th = document.createElement("th");
	th.className = "mob";
	tr.appendChild(th);
	//File Type
	th = document.createElement("th");
	th.innerHTML = "#";
	tr.appendChild(th);

	//FileName
	th = document.createElement("th");
	th.innerHTML = "Name";
	tr.appendChild(th);

	//mob
	th = document.createElement("th");
	th.className = "mob";
	tr.appendChild(th);

	//Size
	th = document.createElement("th");
	th.innerHTML = "Size";
	th.className = "hide-for-small";
	tr.appendChild(th);

	//Mdate
	th = document.createElement("th");
	th.innerHTML = "Change";
	th.className = "hide-for-small";
	tr.appendChild(th);

	//kind
	th = document.createElement("th");
	th.innerHTML = "Kind";
	th.className = "hide-for-small";
	tr.appendChild(th);

	//Actions
	th = document.createElement("th");
	th.innerHTML = "Actions";
	th.className = "hide-for-small";
	tr.appendChild(th);

	this.tbody = document.createElement("tbody");
	browse_div.appendChild(this.thead);
	browse_div.appendChild(this.tbody);

	return browse_div;
}

WualaDisplay.prototype.buildButtonDiv = function(element, displayName, 	onBrowseCB, onDownloadCB, onDeleteCB){
	var buttonDiv = document.createElement("div");
	buttonDiv.className = "actions";
	if (null != onDeleteCB && undefined != onDeleteCB)
	{
		var deleteButton = document.createElement("a");
		deleteButton.className = "button small";
		var i =document.createElement("i");
		i.className = "icon-remove";
		deleteButton.appendChild(i);
		deleteButton.onclick = function(path, event){
			onDeleteCB(event);
		}.bind(a, name);
		buttonDiv.appendChild(deleteButton);
	}


	if (null != onDownloadCB && undefined != onDownloadCB){
		var downloadButton = document.createElement("a");
		downloadButton.className = "button small w-download";
		var i = document.createElement("i");
		i.className = "icon-arrow-down";
		downloadButton.appendChild(i);
		downloadButton.onclick = function(path, event){
			onDownloadCB(event);
		}.bind(a, name);
		buttonDiv.appendChild(document.createTextNode('\u00A0'));
		buttonDiv.appendChild(downloadButton);
	}
	return buttonDiv;
}

WualaDisplay.prototype.AddElement = function(list, element, displayName, onBrowseCB, onDownloadCB, onDeleteCB){
	var name = displayName;
	var tr = document.createElement("tr");
	tr.className = "w-open-path pointer ";
	//TODO TO REMOVE AFTER REWRITING IT
	a = tr;

	//Mobile
	var td = document.createElement("td");
	td.className = "mob";

	var buttonPlus = document.createElement("button");
	buttonPlus.className = "button small w-expand-info";
	var iButtonPlus = document.createElement("i");
	iButtonPlus.className = "icon-plus";
	buttonPlus.appendChild(iButtonPlus);
	buttonPlus.onclick = function(event){
		if (mobtr.style.display == "none"){
			mobtr.style.display = "table-row";
			iButtonPlus.className = "icon-minus";
		}else{
			mobtr.style.display = "none";
			iButtonPlus.className = "icon-plus";
		}
		event.stopPropagation();
	}
	td.appendChild(buttonPlus)
	tr.appendChild(td);

	td=document.createElement("td");
	var img = document.createElement("i");
	img.className = "wred icon-large icon-file-alt";
	td.appendChild(img);
	tr.appendChild(td);

	if (onBrowseCB){
		img.className = "wred icon-large icon-folder-close";
		tr.onclick = function(path, event){
			onBrowseCB(event);
		}.bind(a, name);

	}

	td = document.createElement("td");
	td.innerHTML = name;
	tr.appendChild(td);

	td = document.createElement("td");
	td.className = "mob";
	tr.appendChild(td);

	//Size
	th = document.createElement("td");
	th.innerHTML = "-";
	th.className = "hide-for-small";
	tr.appendChild(th);

	//Mdate
	th = document.createElement("td");
	th.innerHTML = "-";
	th.className = "hide-for-small";
	tr.appendChild(th);

	//kind
	th = document.createElement("td");
	th.innerHTML = "-";
	th.className = "hide-for-small";
	tr.appendChild(th);

	td = document.createElement("td");
	td.className = "hide-for-small";
	tr.appendChild(td);

	var buttonDiv = this.buildButtonDiv(element, displayName, onBrowseCB, onDownloadCB, onDeleteCB);
	td.appendChild(buttonDiv);
	this.tbody.appendChild(a);
	//TODO make another one for the phones :)
	var mobtr = document.createElement("tr");
	mobtr.className = "hide";
	mobtr.style.display = "none";
	td = document.createElement("td");
	td.colSpan = "8";
	mobtr.appendChild(td);

	if (null != onDownloadCB && undefined != onDownloadCB){
		var div = document.createElement("div");
		var strong = document.createElement("strong");
		strong.innerHTML = "Size";
		div.appendChild(strong);
		div.innerHTML += ": " + element.size;
		td.appendChild(div);
		div = document.createElement("div");
		strong = document.createElement("strong");
		strong.innerHTML = "Change";
		div.appendChild(strong);
		div.innerHTML += ": " + element.mDate;
		td.appendChild(div);
		div = document.createElement("div");
		strong = document.createElement("strong");
		strong.innerHTML = "Kind";
		div.appendChild(strong);
		div.innerHTML += ": -";
		td.appendChild(div);
	}
	var actiondiv = this.buildButtonDiv(element, displayName, onBrowseCB, onDownloadCB, onDeleteCB);
	td.appendChild(actiondiv)
	mobtr.appendChild(td);
	//<tr class="hide" data-wuala-path="/CMX10-Android/Android/CM10/Kernel/X10_4.1.X_CMX10_bootmanager_v13.ftf/" style="display: table-row;">
    // <td colspan="8">
    //  <div><strong>Size</strong> : 6.3 MB</div>
    //  <div><strong>Change</strong> : 1 years ago</div>
    //  <div><strong>Kind</strong> :  <i class="muted">ftf</i></div>
    //  <div class="actions">
    //    <a href="https://content.wuala.com/contents/CMX10-Android/Android/CM10/Kernel/X10_4.1.X_CMX10_bootmanager_v13.ftf/?dl=1" class="button small w-download"><i class="icon-arrow-down"></i></a>
    //    <a class="button small w-open-sharing" data-reveal-idd="sharing"><i class="icon-share-alt"></i></a>
    //  </div>
    // </td>
    //</tr>
    this.tbody.appendChild(mobtr);
}

WualaDisplay.prototype.GetListDisplayComponent = function(list){
	return list;
}

WualaDisplay.prototype.GetDeletePopup = function(path, onOkClickCB, onCancelClickCB){

};

WualaDisplay.prototype.GetDownloadPopup = function(path, onDownloadClickCB, onCancelClickCB){

}