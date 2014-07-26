
function ChunkedUploader(file, options) {
    if (!this instanceof ChunkedUploader) {
        return new ChunkedUploader(file, options);
    }

    this.file = file;

    this.options = options;
    this.progressCB = options.progressCB;

    this.file_size = this.file.size;
    this.chunk_size = (1024 * 1024); // 1024KB
    this.range_start = 0;
    this.range_end = this.chunk_size;

    if ('mozSlice' in this.file) {
        this.slice_method = 'mozSlice';
    }
    else if ('webkitSlice' in this.file) {
        this.slice_method = 'webkitSlice';
    }
    else {
        this.slice_method = 'slice';
    }

    this.upload_request = new XMLHttpRequest();
    this.upload_request.onload = this._onChunkComplete.bind(this);
}

ChunkedUploader.prototype = {

// Internal Methods __________________________________________________

    _upload: function() {
        var self = this,
            chunk;

        // Slight timeout needed here (File read / AJAX readystate conflict?)
        setTimeout(function() {
            // Prevent range overflow
            if (self.range_end > self.file_size) {
                self.range_end = self.file_size;
            }

            chunk = self.file[self.slice_method](self.range_start, self.range_end);

            self.upload_request.open('PUT', self.options.url, true);
            if (null != user.authentication_header){
                self.upload_request.setRequestHeader("Authentication", user.authentication_header);
            }
            self.upload_request.setRequestHeader("Content-Type", 'application/octet-stream');

            if (self.range_start !== 0) {
                self.upload_request.setRequestHeader('Content-Range', 'bytes ' + self.range_start + '-' + self.range_end + '/' + self.file_size);
            }
            self.chunk_time = new Date().getTime();
            self.upload_request.send(chunk);

            // TODO
            // From the looks of things, jQuery expects a string or a map
            // to be assigned to the "data" option. We'll have to use
            // XMLHttpRequest object directly for now...
            /*$.ajax(self.options.url, {
                data: chunk,
                type: 'PUT',
                mimeType: 'application/octet-stream',
                headers: (self.range_start !== 0) ? {
                    'Content-Range': ('bytes ' + self.range_start + '-' + self.range_end + '/' + self.file_size)
                } : {},
                success: self._onChunkComplete
            });*/
        }, 20);
    },

// Event Handlers ____________________________________________________

    _onChunkComplete: function() {
        if (this.progressCB){
            this.progressCB(this.range_end);
        }
        console.log("Upload done should carry on");
        // If the end range is already the same size as our file, we
        // can assume that our last chunk has been processed and exit
        // out of the function.
        if (this.range_end === this.file_size) {
            this._onUploadComplete();
            return;
        }

        // Update our ranges
        this.range_start = this.range_end;
        this.range_end = this.range_start + this.chunk_size;

        //Ok check the current time for the response and increase the size of the chunk
        var currentTime = new Date().getTime();
        if ((currentTime - this.chunk_time)/1000 < 1){
            this.chunk_size *= 2;
        }else{
            this.chunk_size /= 2;
        }

        // Continue as long as we aren't paused
        if (!this.is_paused) {
            this._upload();
        }
    },

    _onUploadComplete: function(){
        console.log("Upload completed");
    },

// Public Methods ____________________________________________________

    start: function() {
        this._upload();
    },

    pause: function() {
        this.is_paused = true;
    },

    resume: function() {
        this.is_paused = false;
        this._upload();
    }
};