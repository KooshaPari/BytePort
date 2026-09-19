// Configuration for BytePort Windows deployment
export const config = {
	// Deployment Configuration
	deployment: {
		// Windows-specific deployment settings
		platform: 'windows',
		containerRuntime: 'docker',
		tunnelProvider: 'cloudflare',

		// Default ports for services
		defaultPorts: {
			main: 8080,
			database: 5432,
			redis: 6379
		},

		// Supported project types
		supportedTypes: ['nodejs', 'go', 'python', 'rust', 'static']
	},

	// UI Configuration
	ui: {
		// Polling intervals
		statusPollInterval: 5000, // 5 seconds
		logPollInterval: 2000, // 2 seconds

		// Deployment status messages
		statusMessages: {
			initializing: 'Initializing deployment...',
			building: 'Building Docker container...',
			deploying: 'Starting container...',
			networking: 'Setting up tunnel...',
			completed: 'Deployment completed successfully!',
			failed: 'Deployment failed',
			terminating: 'Terminating project...',
			terminated: 'Project terminated successfully'
		},

		// Windows-specific status indicators
		windowsStatuses: {
			'container-building': 'Building Docker image',
			'container-starting': 'Starting container',
			'tunnel-creating': 'Creating Cloudflare tunnel',
			'tunnel-starting': 'Starting tunnel',
			ready: 'Project is live'
		}
	},

	// Feature flags for Windows deployment
	features: {
		// Enable Windows-specific features
		dockerDeployment: true,
		cloudflareTunnels: true,
		localStorage: true,

		// Disable AWS-specific features
		awsDeployment: false,
		s3Storage: false,
		ec2Instances: false,
		albLoadBalancer: false,

		// Development features
		devMode: import.meta.env.DEV,
		debugLogs: import.meta.env.DEV
	},

	// Windows deployment specific settings
	windows: {
		// Docker settings
		docker: {
			network: 'byteport-network',
			imagePrefix: 'byteport-',
			containerPrefix: 'byteport-'
		},

		// Tunnel settings
		tunnel: {
			configPath: 'C:\\BytePort\\tunnels',
			logPath: 'C:\\BytePort\\logs',
			defaultDomain: 'yourdomain.com'
		},

		// Storage settings
		storage: {
			projectsPath: 'C:\\BytePort\\projects',
			backupsPath: 'C:\\BytePort\\backups'
		}
	}
};

// Export default config
export default config;
