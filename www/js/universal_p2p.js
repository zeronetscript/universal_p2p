'use strict';

function isTorrentFile (file) {
  return file.name.toLowerCase().endsWith('.torrent')
}

function onTorrent(files) {
  console.log('got files');

  if(!isTorrentFile(files[0]) || files.length>1){
    alert("you must only select 1 .torrent file")
    return;
  }
  
  //send form data
  
  $("#torrent_input").prop("files",files)
  $("#torrent_form").submit()


}

var isAdvancedUpload = function() {
  var div = document.createElement('div');
  return (('draggable' in div) || ('ondragstart' in div && 'ondrop' in div)) 
    && 'FormData' in window && 'FileReader' in window;}();

function init(){
  
  $("#hidden_iframe").load(function() {
    alert("complete");
   });

  $("#torrent_input").change(function(){
    $("#torrent_form").submit()
  });

  $("#torrent_form").on("drop",function(e){
    onTorrent(e.originalEvent.dataTransfer.files);
  });



  if (isAdvancedUpload) {
    $("#torrent_form").addClass('has-advanced-upload');
  }
  
}

$(function(){
  init()
});
