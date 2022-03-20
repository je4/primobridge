var log2 = Math.log( 2 );
function nearestPow2( aSize ){
	  
	  var p =  Math.pow( 2, Math.round( Math.log( aSize ) / log2 ) ); 
//	  console.log( "np2: " + aSize + " --> " + p );
	  return p;
}

var Mediathek3D = function( config, done ) {
    // Run the Object3D constructor with the given arguments
    THREE.Object3D.apply(this);

	// function to run if everything is loaded
	this.done = done;
	
	// factor to enhance resolution of Mesh
    this.pixelFactor = 4;

	// size of floor image
    this.floorWidth = 2048;
    this.floorHeight = 2048;
    this.boxWidth = 55;  
    this.boxHeight = 20; 
    this.boxDepth = 30; 
    this.boxThick = 2;
    this.boxDistance = 2; // Distanz zwischen 2 Boxen
    this.boxColor = 0xc0c0c0;
    this.bottomColor = 0xc0c0c0;
    this.boxColorHighlight = 0xff0909;
	this.floorImage = null;
    
    this.satWidth = 80* 2048/6; 
    this.satHeight = 80* 2048/6; 
    this.satImage = null;

	this.boxData = null;
	this.drawOnly = null;
    this.boxDelay = 10;
	
	// overwrite with configuration 
	for(var prop in config)   {
        this[prop]=config[prop];   
    }
	
    
    // make sure that drawOnly is an array
    if ( this.drawOnly != null ) {
        if ( typeof this.drawOnly == 'string') {
            this.drawOnly = [ this.drawOnly ];
        }
    }
    console.log( this.drawOnly );
	// apply the pixelFactor
	this.floorWidth *= this.pixelFactor;
	this.floorHeight *= this.pixelFactor;
    this.satWidth *= this.pixelFactor;
    this.satHeight *= this.pixelFactor;
	this.boxWidth *= this.pixelFactor;
	this.boxHeight *= this.pixelFactor;
	this.boxDepth *= this.pixelFactor;
	this.boxThick *= this.pixelFactor;
	this.boxDistance *= this.pixelFactor;

    this.tangent = new THREE.Vector3();
    this.axis = new THREE.Vector3();
    this.up = new THREE.Vector3(0, 1, 0);
	
	this.floorTexture = null;
    this.satTexture = null;
	this.theData = null;
	this.curves = new Array();
	this.boxes = new Array();
    this.boxNames = new Array();
	this.boxesHighlight = new Array();
	
	// material for normal box
	this.materialBox = new THREE.MeshLambertMaterial({
        color: this.boxColor,
		wireframe: false
    });
	this.materialBottom = new THREE.MeshLambertMaterial({
        color: this.bottomColor,
		wireframe: false
    });
	
	// material for highlighted box
	this.materialBoxHighlight = new THREE.MeshLambertMaterial({
        color: this.boxColorHighlight,
		wireframe: false
    });
	
	this.boxTemplate = this.createBoxTemplate();
	this.bottomTemplate = this.createBottomTemplate();
	
	if( this.floorImage == null ) {
		console.log( "Error: Mediathek3D::floorImage not configured");
		return;
	}
	if( this.boxData == null ) {
		console.log( "Error: Mediathek3D::boxData not configured");
		return;
	}
	
	// load image for floorTexture
	var loader = new THREE.TextureLoader();
	var myself = this;
	loader.load(
		// resource URL
		this.floorImage,
		// Function when resource is loaded
		function ( texture ) {
			// do something with the texture
			myself.floorTexture = new THREE.MeshLambertMaterial( {
				map: texture,
				side: THREE.FrontSide
			 } );
			 myself.loadDone();
		},
		// Function called when download progresses
		function ( xhr ) {
			console.log( (xhr.loaded / xhr.total * 100) + '% loaded' );
		},
		// Function called when download errors
		function ( xhr ) {
			console.log( 'An error happened' );
		}
	);

    if ( this.satImage != null ) {
      // load image for satTexture
      var loader2 = new THREE.TextureLoader();
      var myself2 = this;
      loader2.load(
          // resource URL
          this.satImage,
          // Function when resource is loaded
          function ( texture ) {
              // do something with the texture
              myself.satTexture = new THREE.MeshLambertMaterial( {
                  map: texture,
                  side: THREE.FrontSide
               } );
               myself.satLoadDone();
          },
          // Function called when download progresses
          function ( xhr ) {
              console.log( (xhr.loaded / xhr.total * 100) + '% loaded' );
          },
          // Function called when download errors
          function ( xhr ) {
              console.log( 'An error happened' );
          }
      );
    }
    
	$.getJSON( this.boxData, function( data ) {
//		console.log(JSON.stringify(data));
		myself.theData = data;
		myself.loadDone();
	});
};
// Make Mediathek3D have the same methods as Object3D
Mediathek3D.prototype = Object.create(THREE.Object3D.prototype);
// Make sure the right constructor gets called
Mediathek3D.prototype.constructor = Mediathek3D;
	
// 0 is in the middle of the floor image
Mediathek3D.prototype.calcX = function( x ) {
	return x * this.pixelFactor - (this.floorHeight * 1 /2);
}

Mediathek3D.prototype.calcY = function( y ) {
	return y * this.pixelFactor - (this.floorHeight * 1 /2);
}


// check, whether all needed elements are loaded. if yes, build mediathek
Mediathek3D.prototype.loadDone = function() {
	if( this.floorTexture == null ) return;
	if( this.theData == null ) return;
	
	var floor = new THREE.Mesh(
		new THREE.PlaneGeometry(this.floorHeight * 1, this.floorHeight * 1),
		this.floorTexture
	);
	floor.overdraw = true;
	floor.position.x = 0;
	floor.position.y = 0;
	floor.position.z = 0;
	floor.receiveShadow = true;
	floor.material.side = THREE.DoubleSide;
	this.add(floor);

	this.addBoxes();
   	this.done(this);	
}

Mediathek3D.prototype.satLoadDone = function() {
	if( this.satTexture == null ) return;

	var sat = new THREE.Mesh(
		new THREE.PlaneGeometry(this.satHeight * 1, this.satHeight * 1),
		this.satTexture
	);
	sat.overdraw = true;
	sat.position.y = 0;
	sat.position.z = -2*this.floorHeight;
	sat.receiveShadow = false;
	sat.material.side = THREE.DoubleSide;
	this.add(sat);
}

Mediathek3D.prototype.createBoxTemplate = function() {
	template = new THREE.Object3D();
    material1 = this.materialBox;

	// Boden
	g = new THREE.BoxGeometry(this.boxDepth * 1, 
		this.boxWidth * 1, 
		this.boxThick * 1);
	boxBottom = new THREE.Mesh(g, material1);
	boxBottom.position.set(0,0,0);
	boxBottom.castShadow = true;
	boxBottom.receiveShadow = false;
	boxBottom.name = "bottom";
	template.add(boxBottom);
	// Deckel
	boxTop = new THREE.Mesh(g, material1);
	boxTop.position.set(0,0, this.boxHeight * 1-this.boxThick * 1);
	boxTop.receiveShadow = false
	boxTop.castShadow = true;
	boxTop.name = "top";
	template.add(boxTop);
	// linke Wand
	g = new THREE.BoxGeometry(this.boxDepth * 1, 
		this.boxThick * 1, 
		this.boxHeight * 1-2*this.boxThick * 1);
	boxLeft = new THREE.Mesh(g, material1);
	boxLeft.position.set(0,-(this.boxWidth * 1-this.boxThick * 1)/2, (this.boxHeight * 1-this.boxThick * 1)/2);
	boxLeft.receiveShadow = false
	boxLeft.castShadow = true;
	boxLeft.name = "left";
	template.add(boxLeft);
	// rechte Wand
	boxRight = new THREE.Mesh(g, material1);
	boxRight.position.set(0,(this.boxWidth * 1-this.boxThick * 1)/2, (this.boxHeight * 1-this.boxThick * 1)/2);
	boxRight.receiveShadow = false
	boxRight.castShadow = true;
	boxRight.name = "right";
	template.add(boxRight);
	template.castShadow = true;
	template.receiveShadow = false;
/*    
    var geometry = new THREE.BoxGeometry( 20, 20, 20 );
    var material = new THREE.MeshBasicMaterial( {color: 0x0000ff} );
    var cube = new THREE.Mesh( geometry, material );
    cube.name = "cama";
    cube.position.set( 500*this.pixelFactor, 0, 200*this.pixelFactor );
    template.add( cube );
    var cube = new THREE.Mesh( geometry, material );
    cube.name = "camb";
    cube.position.set( -500*this.pixelFactor, 0, 200*this.pixelFactor );
    template.add( cube );
*/    
	return template;
}

Mediathek3D.prototype.createBottomTemplate = function() {
	template = new THREE.Object3D();
    material1 = this.materialBottom;
    //material1.color = 0xc‚0c0c0;

	// Boden
	g = new THREE.BoxGeometry(this.boxDepth * 1, 
		this.boxWidth * 1, 
		this.boxThick * 1);
	boxBottom = new THREE.Mesh(g, material1);
	boxBottom.position.set(0,0,0);
	boxBottom.castShadow = true;
	boxBottom.receiveShadow = false;
	boxBottom.name = "bottom";
	template.add(boxBottom);

	template.castShadow = true;
	template.receiveShadow = false;
	return template;
}

Mediathek3D.prototype.createBox = function( fullname ) {
	
	var name = fullname.charAt(0);
	var level = fullname.charAt(1);
	var aside = this.theData[name]['aside'];
	var area = this.theData[name]['area'];	
	
	var box = this.boxTemplate.clone();
	box.name = fullname;
	
	var text = "    " + fullname;
	
	var canvas1 = document.createElement('canvas');
	canvas1.width = nearestPow2( this.boxWidth-2*this.boxThick );
	canvas1.height = nearestPow2( this.boxHeight );
	
	context1 = canvas1.getContext('2d');

	context1.fillStyle = "#ff0000";
//		context1.fillRect( 1, 1, canvas1.width-1, canvas1.height -1 );

	context1.font = "Bold " + (6*this.pixelFactor) + "px Arial";
	context1.fillStyle = "black";

	if( aside == 'right' ) {
//		console.log( "aside: right" );
		// Move registration point to the center of the canvas
		context1.translate(canvas1.width/2, canvas1.height/2);

		// Rotate 180 degree
		context1.rotate( Math.PI );

		// Move registration point back to the top left corner of canvas
		context1.translate(-canvas1.width/2, -canvas1.height/2);
	}

	context1.fillText( text, 5*this.pixelFactor, canvas1.height );
//	console.log( text );
	
	var texture1 = new THREE.Texture(canvas1) 
	texture1.needsUpdate = true;
	material1 = new THREE.MeshBasicMaterial( {map: texture1, side:THREE.DoubleSide } );
	material1.transparent = true;
	var mesh1 = new THREE.Mesh(
		new THREE.PlaneGeometry(canvas1.width, canvas1.height),
		material1
	  );
	mesh1.rotateZ( -Math.PI / 2 );
	mesh1.position.set(0,0,this.boxThick+1);
	box.add( mesh1 );

	return box;
}

Mediathek3D.prototype.createBottom = function( fullname ) {
	
	var name = fullname.charAt(0);
	var level = fullname.charAt(1);
	var aside = this.theData[name]['aside'];
	var area = this.theData[name]['area'];	
	
	var box = this.bottomTemplate.clone();
	
	if( true ) {
		box.name = fullname;
		
		var text = "    " + fullname;
		
		var canvas1 = document.createElement('canvas');
		canvas1.width = nearestPow2( this.boxWidth-2*this.boxThick );
		canvas1.height = nearestPow2( this.boxHeight );
		
		context1 = canvas1.getContext('2d');
	
		context1.fillStyle = "#ff0000";
	//		context1.fillRect( 1, 1, canvas1.width-1, canvas1.height -1 );
	
		context1.font = "Bold " + (6*this.pixelFactor) + "px Arial";
		context1.fillStyle = "white";
	
		if( aside == 'right' ) {
	//		console.log( "aside: right" );
			// Move registration point to the center of the canvas
			context1.translate(canvas1.width/2, canvas1.height/2);
	
			// Rotate 180 degree
			context1.rotate( Math.PI );
	
			// Move registration point back to the top left corner of canvas
			context1.translate(-canvas1.width/2, -canvas1.height/2);
		}
	
		context1.fillText( text, 5*this.pixelFactor, canvas1.height );
	//	console.log( text );
		
		var texture1 = new THREE.Texture(canvas1) 
		texture1.needsUpdate = true;
		material1 = new THREE.MeshBasicMaterial( {map: texture1, side:THREE.DoubleSide } );
		material1.transparent = true;
		var mesh1 = new THREE.Mesh(
			new THREE.PlaneGeometry(canvas1.width, canvas1.height),
			material1
		  );
		mesh1.rotateZ( -Math.PI / 2 );
		mesh1.position.set(0,0,this.boxThick+1);
		box.add( mesh1 );
	}

	return box;
}

Mediathek3D.prototype.addBoxes = function() {
	
	for (var label in this.theData) {
		this.addBoxPath( label );
	}
}

Mediathek3D.prototype.addBoxPath = function( name ) {
   var data = this.theData[name];

   var spline = new THREE.CatmullRomCurve3();

 

//    var lineGeometry = new THREE.Geometry();
    var lastPoint = null;
	for( var label in data['point'] ) {
//		console.log( data['point'][label] );
		lastPoint = new THREE.Vector3(this.calcX(data['point'][label]['x']), this.calcY(data['point'][label]['y']), 1);
		spline.points.push( lastPoint );
        //lineGeometry.vertices.push(new THREE.Vector3(this.calcX(data['point'][label]['x']), this.calcY(data['point'][label]['y']), 1));
	}
	
/*
	var material = new THREE.LineBasicMaterial({
        color: 0xff00f0,
    });
	var line = new THREE.Line(lineGeometry, material);
    this.add(line);
*/	
	this.curves[name] = spline;

	
	for( var level in data['level'] ) {
		this.addBoxLevel(name, level );
	}
}	

Mediathek3D.prototype.addBoxLevel = function( name, level ) {
	var spline = this.curves[name];
	var aside = this.theData[name]['aside'];
	var area = this.theData[name]['area'];
	var data = this.theData[name]['level'][level]
	
	indent = data['indent'];
	splineLength = spline.getLength();
//	console.log( "Spline length: " + splineLength);
	
	
	for( counter = 0; counter < data['boxes']; counter ++) {
		
		boxname = name + level + (counter < 9 ? '0' : '') + (counter+1);
		if( this.drawOnly == null ) 
			box = this.createBox( boxname );
		else {
            var found = false;
            for( var i = 0; i < this.drawOnly.length; i++ ) {
                  if( this.drawOnly[i].substring( 0,1 ) == name ) {
                        found = true;
                        break;
                  }
            }
            if ( found ) {
                  box = this.createBox( boxname );
            }
            else {
                if( level == 1 )
                    box = this.createBottom( boxname );
                else
                    return;
            }
		}
			 
		
		//p = counter*(1/10)+1/20;
		p = ((counter+indent) * (this.boxWidth + this.boxDistance) + this.boxWidth/2);
//		console.log( "Spline pos: " + p);
		p /=  splineLength;
		
		// set base position
		box.position.copy( spline.getPointAt(p) );
        
		// get normalized tangent vector
		tangent = spline.getTangentAt(p).normalize();
        
		// Kreuzprodukt
		this.axis.crossVectors(this.up, tangent).normalize();
        
		// Skalarprodukt)
		var radians = Math.acos(this.up.dot(tangent));
        
		box.quaternion.setFromAxisAngle(this.axis, radians);		
		
		box.position.z = (level-1) * this.boxHeight + this.boxThick;
		box.name = boxname;
		this.boxes[boxname] = box;
        this.boxNames.push( boxname );
		//this.add(box);	
	}
}

Mediathek3D.prototype.renderBoxes = function( id ) {
      if ( this.boxNames.length <= id ) {
            return;
      }
      this.add( this.boxes[this.boxNames[id]] );
      var _this = this;
      setTimeout( function() { _this.renderBoxes( id+1); }, this.boxDelay );
}

Mediathek3D.prototype.boxHighlight = function( boxname, highlight ) {
	console.log( "Highlight( " + boxname + ", " + highlight + ")");
	if(!( boxname in this.boxes )) {
		console.log( "box " + boxname + " does not exist");
		return;
	}
	
	var box = this.boxes[boxname];
	var material = highlight ? this.materialBoxHighlight : this.materialBox;
	box.traverse( function(child) {
		switch( child.name ) {
		case "top":
		case "bottom":
		case "left":
		case "right":
			child.material = material;
		}
	});
	if( boxname in this.boxesHighlight ) {
		delete this.boxesHighlight[boxname];
	}
	if( highlight ) {
		this.boxesHighlight[boxname] = true;
	}
}

Mediathek3D.prototype.clearHighlight = function() {
	for( var key in this.boxesHighlight ) {
		this.boxHighlight(key, false );
	}
}

