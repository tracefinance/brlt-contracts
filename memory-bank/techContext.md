# Vault0 Technical Context

## Technologies Used

### Backend Technologies

1. **Go (Golang)**
   - Primary backend language
   - Version: Latest stable (1.16+)
   - Used for all server-side logic

2. **SQLite**
   - Database system
   - File-based relational database
   - Used for persistent storage

3. **Gin Web Framework**
   - HTTP framework for Go
   - Used for REST API implementation
   - Handles routing, middleware, and HTTP interactions

4. **OAuth2**
   - Authentication framework
   - Implementation for secure user authentication
   - Integration with web and mobile clients

5. **Cryptography Libraries**
   - AES-GCM for encryption
   - secp256k1 for key generation and management
   - Used for secure key storage and blockchain operations

### Smart Contract Technologies

1. **Solidity**
   - Smart contract programming language
   - Version: ^0.8.28
   - Used for implementing MultiSigWallet contract

2. **Hardhat**
   - Development environment for Ethereum
   - Used for testing, compilation, and deployment
   - Provides local blockchain for development

3. **OpenZeppelin Contracts**
   - Library of secure, reusable smart contract components
   - Used for standard implementations and security patterns
   - Provides ERC20 interfaces and utility functions

### Frontend Technologies

1. **Next.js**
   - React framework for web applications
   - Version: 15.2
   - Server-side rendering and static site generation

2. **React**
   - JavaScript library for building user interfaces
   - Version: 19
   - Component-based UI development

3. **TypeScript**
   - Typed JavaScript
   - Used for all frontend code
   - Provides type safety and improved development experience

4. **TailwindCSS**
   - Utility-first CSS framework
   - Version: 4
   - Used for responsive design and styling

## Development Setup

### Prerequisites

- **Node.js**: ≥ 14.0.0
- **npm**: ≥ 6.0.0
- **Go**: ≥ 1.16
- **SQLite**: Latest version
- **Git**: For version control

### Backend Setup

1. **Environment Variables**:
   - `DB_ENCRYPTION_KEY`: Required for database encryption
   - `JWT_SECRET`: Used for JWT token signing
   - `DB_PATH`: Path to SQLite database file

2. **Database Setup**:
   - Migrations located in `/migrations`
   - Applied automatically on server start
   - Can be manually applied using the `go-migrate` tool

3. **Key Generation**:
   - Use the provided `genkey` utility to generate encryption keys
   - Example: `./bin/genkey`

4. **Development Server**:
   - Start with `make server-dev`
   - Runs on `localhost:8080` by default
   - API documentation available at `/swagger/index.html`

### Smart Contract Setup

1. **Environment Variables**:
   - `.env` file in `/contracts` directory
   - `PRIVATE_KEY`: Deployer wallet private key
   - `BASE_RPC_URL`: Base network RPC URL
   - `POLYGON_ZKEVM_RPC_URL`: Polygon zkEVM RPC URL

2. **Deployment**:
   - Use Hardhat scripts for deployment
   - Different commands for different networks
   - Example: `make contracts-deploy-base-test`

3. **Testing**:
   - Run test suite with `make contracts-test`
   - Coverage reports with `make contracts-test-coverage`
   - Requires 90%+ coverage for production

### Frontend Setup

1. **Environment Variables**:
   - `.env.local` file in `/ui` directory
   - `NEXT_PUBLIC_API_URL`: Backend API URL
   - `NEXT_PUBLIC_CHAIN_IDS`: Supported blockchain network IDs

2. **Development Server**:
   - Start with `make ui-dev`
   - Runs on `localhost:3000` by default

3. **Build for Production**:
   - Build with `make ui`
   - Start production server with `make ui-start`

## Technical Constraints

### Blockchain Limitations

1. **Network Support**:
   - Limited to EVM-compatible networks
   - Specifically Base and Polygon zkEVM
   - Cannot support non-EVM chains without significant rework

2. **Gas Considerations**:
   - Operations must be optimized for gas efficiency
   - Complex operations limited by block gas limits
   - Gas costs directly impact user experience

3. **Smart Contract Immutability**:
   - Deployed contracts cannot be modified
   - Requires careful testing and auditing before deployment
   - Upgrade patterns needed for future improvements

### Security Requirements

1. **Key Management**:
   - Private keys must never be exposed
   - All keys must be encrypted at rest
   - Encryption key must be provided via environment variable

2. **Authentication**:
   - All API endpoints must be properly secured
   - JWT tokens with appropriate expiration
   - Refresh token mechanism for extended sessions

3. **Input Validation**:
   - All inputs must be validated at API boundary
   - Smart contract functions must validate parameters
   - Defense in depth approach to security

### Performance Considerations

1. **Database Performance**:
   - SQLite limitations for concurrent access
   - Appropriate indexing for common queries
   - Connection pooling for efficient resource usage

2. **Blockchain Interaction**:
   - RPC rate limiting by providers
   - Transaction confirmation times
   - Event monitoring efficiency

3. **Frontend Responsiveness**:
   - Optimized bundle size
   - Efficient React component rendering
   - Progressive loading strategies

## Dependencies

### Backend Dependencies

1. **Core Dependencies**:
   - `github.com/gin-gonic/gin`: Web framework
   - `github.com/mattn/go-sqlite3`: SQLite driver
   - `golang.org/x/crypto`: Cryptography functions
   - `github.com/ethereum/go-ethereum`: Ethereum client

2. **Authentication**:
   - `github.com/golang-jwt/jwt`: JWT implementation
   - `github.com/go-oauth2/oauth2`: OAuth2 server

3. **Utilities**:
   - `github.com/google/uuid`: UUID generation
   - `go.uber.org/zap`: Structured logging
   - `github.com/spf13/viper`: Configuration management

### Smart Contract Dependencies

1. **Core Dependencies**:
   - `@openzeppelin/contracts`: Security-audited contract components
   - `hardhat`: Development environment
   - `ethers`: Ethereum library

2. **Testing**:
   - `chai`: Assertion library
   - `mocha`: Test framework
   - `solidity-coverage`: Code coverage for Solidity

3. **Deployment**:
   - `hardhat-deploy`: Deployment scripting
   - `@nomiclabs/hardhat-etherscan`: Contract verification

### Frontend Dependencies

1. **Core Dependencies**:
   - `next`: React framework
   - `react`: UI library
   - `typescript`: Type system

2. **UI Components**:
   - `tailwindcss`: CSS framework
   - `@headlessui/react`: Accessible UI components
   - `react-icons`: Icon library

3. **Blockchain Integration**:
   - `ethers`: Ethereum library
   - `wagmi`: React hooks for Ethereum
   - `@web3modal/react`: Wallet connection UI

4. **Data Management**:
   - `swr`: Data fetching and caching
   - `axios`: HTTP client
   - `react-query`: Data fetching library
