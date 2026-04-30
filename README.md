# SkyFox Cinemas
SkyFox is a comprehensive, scalable, and robust web-based platform that modernizes cinema management while delivering a seamless, secure, and user-friendly ticket booking experience

## Backend
Golang & Gin HTTP framework

## Frontend
ReactJS

## DB
PostgreSQL

##  Features of SkyFox Cinema
- Microservices Architecture: Decoupled system design with independent services for Movies and Bookings, allowing for better scalability and maintenance.

- Real-time Movie Scheduling: Backend logic in stormborne-booking-backend to manage showtimes, theater assignments, and movie metadata updates.

- Secure Seat Reservation: A transactional booking system that prevents double-booking and manages seat availability in real-time.

- RBAC with JWT Authentication: Implements JSON Web Tokens (JWT) to manage Role-Based Access Control, ensuring that only authorized users (e.g., Admins) can access sensitive endpoints like revenue data or schedule management.

- Admin Analytics Dashboard: High-level monitoring tools for tracking revenue, ticket sales, and platform performance data.

- QR Code Check-in System: Digital ticketing implementation allowing for on-the-ground efficiency and rapid customer entry verification.

- Automated Database Management: Structured migration scripts and seeding tools to ensure consistent database schemas across development and production.

- Robust Error Handling & Logging: Centralized middleware for request validation, CORS management, and detailed system logging.

- Frontend: A high-performance ReactJS interface optimized for fast loading and smooth user transitions.

- Integrated Testing Suite: Extensive use of mocks and integration tests for controllers, services, and repositories to ensure code reliability.