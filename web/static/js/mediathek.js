
var Mediathek = function( config, done ) {
	this.controltype = 'trackball';
	this.camType = "perspective";
	this.renderDOM = null;
	this.backgroundColor = 0xFAFAFA;
	this.pixelFactor = 8;
	this.floorWidth = 2048;
    this.floorHeight = 2048;
    this.ambientLightColor = 0x404040;
	this.doWindowResize = false;
//	this.showSatellite = true;
    this.camJSON = null; //'[0.9993451833724976,0.015195777639746666,-0.03283816948533058,0,0.01806975156068802,0.5766857266426086,0.8167662024497986,0,0.03134870156645775,-0.816824734210968,0.5760335326194763,0,286.9632568359375,-7477.142578125,5272.96044921875,1]';
	this.animationID = null;
	
	// overwrite with configuration 
	for(var prop in config)   {
        this[prop]=config[prop];   
    }
	
	if( this.renderDOM == null ) {
		this.renderDOM = $(document.body);
	}
	
	this.scene = new THREE.Scene();;
	this.renderer = null;
	this.createRenderer();

	this.camera = null;
	switch( this.camType ) {
		case "perspective":
			this.createPerspectiveCamera();
			break;
		case "ortho":
			this.createOrthoCamera();
			break;
		default:
			console.log( "Error: no camera created" );
		return;
	}
	this.controls = null;
	this.createControls();
	
	this.addLight();
	
	if( this.camJSON != null )
		this.setCamJSON( this.camJSON );
	
	if( this.doWindowResize )
		this.windowResize = THREEx.DOMResize( this.renderer, this.camera, this.renderDOM );
}

Mediathek.prototype.getCamJSON = function() {
	if( this.camera == null ) return "";
	var state = {position: ( this.camera.position.toArray()),
					up: ( this.camera.up.toArray()),
					target: ( this.controls.target.toArray()),
			}	
	return JSON.stringify( state );
}

Mediathek.prototype.setCamJSON = function( jsonstr ) {
	var state = JSON.parse( jsonstr );
	this.camera.position.set( state.position[0], state.position[1], state.position[2] );
	this.camera.up.set( state.up[0], state.up[1], state.up[2] );
	this.controls.target.set( state.target[0], state.target[1], state.target[2] ); 
}

Mediathek.prototype.createRenderer = function() {
    this.renderer = new THREE.WebGLRenderer({ alpha: true,
											  preserveDrawingBuffer: true });
    this.renderer.setSize(this.renderDOM.innerWidth(), this.renderDOM.innerHeight());
    this.renderer.shadowMap.enabled = true;
    this.renderer.setClearColor( this.backgroundColor, 1);
	//renderer.shadowMapSoft = true;
    this.renderer.shadowMap.type = THREE.PCFSoftShadowMap;
	copy = document.createElement("div");
	copy.appendChild( document.createTextNode( "(c) copyright 2016 info-age GmbH Basel"));
	this.renderer.domElement.appendChild( copy );
		this.renderDOM.append(this.renderer.domElement);	
}

Mediathek.prototype.createPerspectiveCamera = function() {
		this.camera = new THREE.PerspectiveCamera(45, this.renderDOM.innerWidth() / this.renderDOM.innerHeight(), 1, 5000 * this.pixelFactor);
	    this.camera.position.set(0 * this.pixelFactor, -900 * this.pixelFactor, 300 * this.pixelFactor);
		// top view
	    this.camera.position.set(0 * this.pixelFactor, 0 * this.pixelFactor, 900 * this.pixelFactor); 
    this.camera.lookAt(new THREE.Vector3(this.floorWidth/2, this.floorHeight/2, 0));
	
}

Mediathek.prototype.createOrthoCamera = function() {
	var aspectRatio = this.renderDOM.innerWidth() / this.renderDOM.innerHeight();
	
	this.camera = new THREE.OrthographicCamera( 
				0.5 * 2048*this.pixelFactor*aspectRatio/-2, 
				0.5 * 2048*this.pixelFactor*aspectRatio/ 2, 
				0.5 * 2048*this.pixelFactor/ 2, 
				0.5 * 2048*this.pixelFactor/-2, 
				-2048*this.pixelFactor, 
				2048*this.pixelFactor );
	//this.camera.position.set(0 * this.pixelFactor, -0 * this.pixelFactor, 500 * this.pixelFactor);
	//var cameraOrthoHelper = new THREE.CameraHelper( this.camera );
	//this.scene.add( cameraOrthoHelper );
/*	
	this.camera = new THREE.OrthographicCamera( this.renderDOM.innerWidth / -16, this.renderDOM.innerWidth / 16, 
		this.renderDOM.innerHeight / 16, this.renderDOM.innerHeight / -16, 
		-200*this.pixelFactor, 2000*this.pixelFactor );
*/		
//	this.camera.position.set(24 * this.pixelFactor, -2410 * this.pixelFactor, 1294 * this.pixelFactor);
//	this.camera.lookAt(new THREE.Vector3(this.floorWidth/2, this.floorHeight/2, 0));
}

Mediathek.prototype.windowResize = function() {
	if( this.camera == null ) return;

	
    this.camera.aspect = this.renderDOM.innerWidth() / this.renderDOM.innerHeight();
//    this.camera.setSize(this.renderDOM.innerWidth(), this.renderDOM.innerHeight());
	this.camera.updateProjectionMatrix();

    this.renderer.setSize( this.renderDOM.innerWidth()-2, this.renderDOM.innerHeight()-2 );
//	console.log( "resize" );
	
}

Mediathek.prototype.addLight = function() {
	var ambientLight = new THREE.AmbientLight( this.ambientLightColor );
	this.scene.add( ambientLight );	

	var directionalLight = new THREE.DirectionalLight(0x606060, 1);
	directionalLight.position.set(-this.floorWidth/8 * this.pixelFactor, -this.floorWidth/8 * this.pixelFactor, 1000 * this.pixelFactor);
	directionalLight.target.position.set(this.floorWidth/4 * this.pixelFactor, this.floorWidth/4 * this.pixelFactor, 0);
	directionalLight.castShadow = true;
	directionalLight.shadow.darkness = 100;
	 
	directionalLight.shadow.camera.near = 10*this.pixelFactor;
	directionalLight.shadow.camera.far = 4000 * this.pixelFactor;
	 
	directionalLight.shadow.camera.left = 2048 * this.pixelFactor /-4;
	directionalLight.shadow.camera.right = 2048 * this.pixelFactor / 4;
	directionalLight.shadow.camera.top = 2048 * this.pixelFactor / 4;
	directionalLight.shadow.camera.bottom = 2048 * this.pixelFactor /-4;	
	this.scene.add( directionalLight );
	
//	var directionalLightHelper = new THREE.DirectionalLightHelper( directionalLight, 200*this.pixelFactor );
//	directionalLightHelper.update();
//	this.scene.add( directionalLightHelper );
//	var helper = new THREE.CameraHelper( directionalLight.shadow.camera );
//	this.scene.add( helper );
	
 
	directionalLight = new THREE.DirectionalLight(0x606060, 1);
	directionalLight.position.set(-this.floorWidth/4 * this.pixelFactor, this.floorWidth/4 * this.pixelFactor, 2500 * this.pixelFactor);
	directionalLight.target.position.set(this.floorWidth/4 * this.pixelFactor, -this.floorWidth/4 * this.pixelFactor, 0);
	directionalLight.castShadow = false;
	this.scene.add( directionalLight );
//	directionalLightHelper = new THREE.DirectionalLightHelper( directionalLight, 200*this.pixelFactor );
//	directionalLightHelper.update();
//	this.scene.add( directionalLightHelper );
	return;

	directionalLight = new THREE.DirectionalLight(0xa0a0a0, 1);
	directionalLight.position.set(this.floorWidth/4 * this.pixelFactor, this.floorWidth/4 * this.pixelFactor, 1500 * this.pixelFactor);
	directionalLight.target.position.set(-this.floorWidth/4 * this.pixelFactor, -this.floorWidth/4 * this.pixelFactor, 0);
	directionalLight.castShadow = false;
	this.scene.add( directionalLight );
	directionalLightHelper = new THREE.DirectionalLightHelper( directionalLight, 200*this.pixelFactor );
	directionalLightHelper.update();
	this.scene.add( directionalLightHelper );

	directionalLight = new THREE.DirectionalLight(0xa0a0a0, 1);
	directionalLight.position.set(this.floorWidth/4 * this.pixelFactor, -this.floorWidth/4 * this.pixelFactor, 1500 * this.pixelFactor);
	directionalLight.target.position.set(-this.floorWidth/4 * this.pixelFactor, this.floorWidth/4 * this.pixelFactor, 0);
	directionalLight.castShadow = false;
	this.scene.add( directionalLight );
	directionalLightHelper = new THREE.DirectionalLightHelper( directionalLight, 200*this.pixelFactor );
	directionalLightHelper.update();
	this.scene.add( directionalLightHelper );

	return;
}

Mediathek.prototype.createControls = function () {
	if ( this.controltype == 'fly') {
        this.controls = new THREE.FlyControls( this.camera, this.renderer.domElement );
		this.controls.movementSpeed = 0.05;
		this.controls.rollSpeed = Math.PI / 400;
		this.controls.autoForward = true;
		this.controls.dragToLook = false;

    }
	else {
		this.controls = new THREE.TrackballControls(this.camera, this.renderer.domElement);
	//	this.controls = new THREE.OrbitControls( this.camera, this.renderer.domElement );
	
		this.controls.rotateSpeed = 2.0;
		this.controls.zoomSpeed = 1.2;
		this.controls.panSpeed = 0.8;
	
		this.controls.noZoom = false;
		this.controls.noPan = false;
	
		this.controls.staticMoving = true;
		this.controls.dynamicDampingFactor = 0.3;
	
	//	this.controls.keys = [ 65, 83, 68 ];	
	}
}

Mediathek.prototype.render = function() {
    this.controls.update(1);
//	directionalLight.position.copy( camera.position );
    this.renderer.render(this.scene, this.camera);
/*	
	// camera looking down internal z-axis
	var point = new THREE.Vector3( 0, 0, -1 );
	// convert from camera space to world-space
	point.applyMatrix4( camera.matrixWorld );
	console.log( "Camera: " + camera.position.x + "/" + camera.position.y + "/" + camera.position.z )
	console.log( "Camera look: " + point.x + "/" + point.y + "/" + point.z )
*/	
}

Mediathek.prototype.animate = function() {
    mediathek.render();
	var that = this;
    this.animationID = requestAnimationFrame(function () {
		that.animate();
	});
}

Mediathek.prototype.stopAnimate = function() {
	if ( this.animationID ) {
	    cancelAnimationFrame( this.animationID );
		this.animationID = null;
    }
}
