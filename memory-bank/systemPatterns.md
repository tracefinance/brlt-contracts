# Vault0 System Patterns

## System Architecture

The Vault0 project follows a comprehensive multi-layer architecture across its three main components: smart contracts, backend services, and frontend application.

### Three-Layer Backend Architecture

The Go backend implements a clear separation of concerns through a three-layer architecture:

```mermaid
flowchart TD
    A[Communication Layer] --> B[Service Layer]
    B --> C[Core/Infrastructure Layer]
    C --> D[External Systems]
    
    subgraph "Layer 3: Communication"
        A1[API Handlers]
        A2[Middleware]
        A3[Routes]
    end
    
    subgraph "Layer 2: Services"
        B1[Business Logic]
        B2[Data Models]
        B3[Repositories]
    end
    
    subgraph "Layer 1: Core"
        C1[Blockchain]
        C2[Keystore]
        C3[Wallet]
        C4[Crypto]
        C5[Database]
        C6[Contracts]
    end
    
    subgraph "External"
        D1[Blockchain Networks]
        D2[Database Storage]
    end
    
    A --> A1 & A2 & A3
    B --> B1 & B2 & B3
    C --> C1 & C2 & C3 & C4 & C5 & C6
    D --> D1 & D2
```

1. **Layer 1: Core/Infrastructure**
   - Located in `internal/core/`
   - Provides fundamental building blocks
   - Interfaces with external systems
   - Handles low-level operations

2. **Layer 2: Service Layer**
   - Located in `internal/services/`
   - Implements business logic
   - Domain-specific operations
   - Data models and repositories

3. **Layer 3: Communication Layer**
   - Located in `internal/api/`
   - Exposes functionality through REST endpoints
   - Manages request/response handling
   - Implements middleware for cross-cutting concerns

### Smart Contract Architecture

The smart contract component centers around the MultiSigWallet contract:

```mermaid
flowchart TD
    MSW[MultiSigWallet Contract] --> ERC20[ERC20 Interface]
    MSW --> Events[Event System]
    MSW --> TimeL[Timelock System]
    MSW --> Whitelist[Token Whitelist]
    MSW --> Recovery[Recovery Mechanism]
    MSW --> DualSig[Dual Signature Verification]
```

### Frontend Architecture

The Next.js frontend follows a component-based architecture:

```mermaid
flowchart TD
    Pages[Pages] --> Layouts[Layout Components]
    Pages --> Domain[Domain Components]
    Domain --> Common[Common Components]
    Pages --> Hooks[React Hooks]
    Hooks --> API[API Client]
    API --> Backend[Backend API]
```

## Key Technical Decisions

### 1. Go for Backend Development

**Decision**: Implement the backend in Go rather than Node.js or other alternatives.

**Rationale**:
- Performance benefits for cryptographic operations
- Strong type system reduces runtime errors
- Excellent concurrency model
- Compiles to a single binary for easy deployment
- Good support for blockchain integration

### 2. Multi-Layer Architecture

**Decision**: Structure the backend in three distinct layers.

**Rationale**:
- Clear separation of concerns
- Improved testability
- Better maintainability
- Flexibility to change implementations

### 3. Interface-Driven Development

**Decision**: Define interfaces before implementations across the system.

**Rationale**:
- Enables dependency injection
- Facilitates testing through mocks
- Allows multiple implementations
- Decouples components

### 4. SQLite Database

**Decision**: Use SQLite for data persistence.

**Rationale**:
- Simplifies deployment (no separate database server)
- Sufficient performance for expected load
- Reliable and well-tested
- File-based storage fits the application needs

### 5. Dual-Signature Security Model

**Decision**: Require two signatures (client and manager) for withdrawals.

**Rationale**:
- Eliminates single point of failure
- Implements separation of duties
- Provides appropriate security for high-value operations
- Balances security with usability

### 6. 72-hour Timelock Recovery

**Decision**: Implement a 72-hour timelock for recovery operations.

**Rationale**:
- Provides sufficient time to detect unauthorized recovery attempts
- Balances security with practical recovery needs
- Long enough for intervention but not excessively delayed

## Design Patterns

### Repository Pattern

Used throughout the service layer to abstract data access:

```go
// Repository interface example
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id uuid.UUID) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### Factory Pattern

Used for creating blockchain, contract, and wallet instances:

```go
// Factory example
func NewBlockchain(ctx context.Context, config *BlockchainConfig) (Blockchain, error) {
    switch config.Type {
    case "evm":
        return NewEvmBlockchain(ctx, config)
    default:
        return nil, fmt.Errorf("unsupported blockchain type: %s", config.Type)
    }
}
```

### Dependency Injection

Used throughout the application to provide dependencies:

```go
// Service constructor with injected dependencies
func NewUserService(
    repo UserRepository, 
    keystore Keystore,
    logger *logger.Logger,
) *UserService {
    return &UserService{
        repo:     repo,
        keystore: keystore,
        logger:   logger,
    }
}
```

### Middleware Pattern

Used in the API layer for cross-cutting concerns:

```go
// Middleware example
func AuthMiddleware(authService *AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
                Error: "missing authentication token",
            })
            return
        }
        
        userID, err := authService.ValidateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
                Error: "invalid authentication token",
            })
            return
        }
        
        c.Set("userID", userID)
        c.Next()
    }
}
```

### Service Pattern

Used to encapsulate business logic:

```go
// Service example
type WalletService struct {
    repo       WalletRepository
    blockchain Blockchain
    keystore   Keystore
    logger     *zap.Logger
}

func (s *WalletService) CreateWallet(ctx context.Context, userID uuid.UUID, chainID uint64) (*Wallet, error) {
    // Business logic implementation
}
```

## Component Relationships

### Backend Components

```mermaid
flowchart TD
    API[API Layer] --> US[User Service]
    API --> WS[Wallet Service]
    API --> TS[Transaction Service]
    API --> BS[Blockchain Service]
    
    US --> UR[User Repository]
    WS --> WR[Wallet Repository]
    WS --> K[Keystore]
    WS --> BC[Blockchain]
    TS --> TR[Transaction Repository]
    TS --> BC
    BS --> BC
    
    BC --> C[Contract]
    BC --> W[Wallet]
    
    UR --> DB[(Database)]
    WR --> DB
    TR --> DB
    K --> DB
    
    BC --> BN[Blockchain Networks]
```

### Smart Contract Components

```mermaid
flowchart TD
    MSW[MultiSigWallet] --> WL[Withdrawal Logic]
    MSW --> RL[Recovery Logic]
    MSW --> TL[Token Logic]
    MSW --> SL[Security Logic]
    
    WL --> DSV[Dual Signature Verification]
    WL --> EXP[Expiration Check]
    RL --> TLK[Timelock]
    TL --> WHT[Whitelist Check]
    SL --> REE[Reentrancy Protection]
    SL --> ACC[Access Control]
```

### User Authentication Flow

```mermaid
flowchart TD
    Login[Login Request] --> Auth[Authentication]
    Auth --> JWT[JWT Token Generation]
    JWT --> Client[Return to Client]
    
    API[API Request] --> Verify[Verify Token]
    Verify --> Valid{Valid?}
    Valid -->|Yes| Process[Process Request]
    Valid -->|No| Reject[Reject Request]
```

### Transaction Flow

```mermaid
flowchart TD
    Client[Client] --> Init[Initiate Withdrawal]
    Init --> Store[Store in Database]
    Store --> Notify[Notify Manager]
    
    Manager[Manager] --> Review[Review Withdrawal]
    Review --> Approve{Approve?}
    
    Approve -->|Yes| Sign[Sign Transaction]
    Approve -->|No| Reject[Reject Transaction]
    
    Sign --> Submit[Submit to Blockchain]
    Submit --> Verify[Verify Dual Signatures]
    Verify --> Execute[Execute Transfer]
