'use strict';
var parseTorrentFile = require('parse-torrent-file')

function isTorrentFile (file) {
  return file.name.toLowerCase().endsWith('.torrent')
}

function submitTorrentComplete(response){

  $("#command_log").append(response)


}

function receiveMessage(event)
{

}

function onTorrent(files) {
  console.log('got files');

  if(!isTorrentFile(files[0]) || files.length>1){
    alert("you must only select 1 .torrent file");
    return;
  }
  
  //send form data
  var data = new FormData();
  data.append("UPLOAD",files[0]);

  $.ajax({
    url: window.location.origin+"/bittorrent/v0/add_torrent",
    data:data,
    cache:false,
    contentType:false,
    processData:false,
    type:'POST',
    success:submitTorrentComplete
  });

}

var isAdvancedUpload = function() {
  var div = document.createElement('div');
  return (('draggable' in div) || ('ondragstart' in div && 'ondrop' in div)) 
    && 'FormData' in window && 'FileReader' in window;}();

function init(){

  $("html").on("dragover", function(event) {
    event.preventDefault();  
    event.stopPropagation();
  });

  $("html").on("dragleave", function(event) {
    event.preventDefault();  
    event.stopPropagation();
  });

  $("html").on("drop", function(event) {
    event.preventDefault();  
    event.stopPropagation();
  });

  
  $("#hidden_iframe").load(function() {
    var p = $(this).contentDocument
    parent.postMessage("")
   });

  $("#torrent_input").change(function(){
    onTorrent($("#torrent_input").prop("files"));
  });




  [ 'drag', 'dragstart', 'dragend', 'dragover', 'dragenter', 'dragleave', 'drop' ].forEach( function( event )
      {
        $("#torrent_form").bind( event, function( e )
            {
              // preventing the unwanted behaviours
              e.preventDefault();
              e.stopPropagation();
            });
      });
  [ 'dragover', 'dragenter' ].forEach( function( event )
      {
        $("#torrent_form").bind( event, function()
            {
              $("#torrent_form").addClass( 'is-dragover' );
            });
      });
  [ 'dragleave', 'dragend', 'drop' ].forEach( function( event )
      {
        $("#torrent_form").bind( event, function()
            {
              $("#torrent_form").removeClass( 'is-dragover' );
            });
      });

  $("#torrent_form").bind( 'drop', function( e )
      {
        $("#torrent_input").prop("files",e.originalEvent.dataTransfer.files);
      });


  if (isAdvancedUpload) {
    $("#torrent_form").addClass('has-advanced-upload');
  }

  window.addEventListener("message",receiveMessage,false);

}

$(function(){
  init()
});
