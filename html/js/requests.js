
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
        	if ((null != obj.onError) && (undefined != obj.onError)){
        		obj.onError(request, status, error);
        	}
        },
        dataType:"json"
    }
    if (null != obj.data || undefined != obj.data){
    	request.data = JSON.stringify(obj.data);
    }
   	$.ajax(request);
}

function sendCommand(request){
	request.url = "commands";
	request.method = "POST";
	var key = queryString["key"];
	if (undefined != key){
		request.data.auth_key = key;
	}
	if (undefined != request.poll && request.poll){
		var tempOnSuccess = request.onSuccess;
		request.onSuccess = function(result){
			if (2 == result.state.status){
				window.setTimeout(function(){
					sendRequest({
						url:"commands/" + this.command_id,
						method:"GET",
						onSuccess: tempOnSuccess
					});
				}.bind(result), 3000);
				return
			}
			tempOnSuccess(result);
		}
	}
	sendRequest(request);
}