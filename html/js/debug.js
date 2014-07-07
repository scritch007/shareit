function Notification(options){
	var notification = document.createElement("div");
	var notificationLabel = document.createElement("label");
	notificationLabel.innerHTML = options.name;
	notification.appendChild(notificationLabel);
	if (options.progressBar){
		var progressBar = document.createElement("progress");
		progressBar.max = "100";
		progressBar.value = "0";
		notification.appendChild(progressBar);
		notification.progressBar = progressBar;
	}
	return notification;
}

function notification(){
	var notification_div = document.getElementById("notifications");
	var newNotification = new Notification({name:"test"});
	notification_div.appendChild(newNotification);
	var value = 0;
	function updateMe(){
		newNotification2.progressBar.value = value++;
		if (100 <= value){
			window.clearInterval(intervalVariable);
		}
	}
	var newNotification2 = new Notification({name:"progresbar", progressBar:true});
	var intervalVariable = setInterval(function(){updateMe();}, 100);
	notification_div.appendChild(newNotification2);
}