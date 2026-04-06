export default function ProfilePage() {
  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Profile</h1>
          <p>View and manage your account information.</p>
        </div>
      </div>

      <section className="dashboard-cards">
        <div className="dashboard-card dashboard-card-primary">
          <h3>Profile</h3>
          <p>View and manage your account information.</p>
        </div>

        <div className="dashboard-card">
          <h3>Account</h3>
          <p>Inspect personal details and authentication data.</p>
        </div>

        <div className="dashboard-card">
          <h3>Security</h3>
          <p>Review password and access-related settings.</p>
        </div>
      </section>
    </>
  )
}