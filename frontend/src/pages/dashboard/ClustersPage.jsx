export default function ClustersPage() {
  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Clusters</h1>
          <p>Inspect connected Kubernetes clusters and their status.</p>
        </div>
      </div>

      <section className="dashboard-cards">
        <div className="dashboard-card dashboard-card-primary">
          <h3>Connected clusters</h3>
          <p>Cluster data from your deployed agents will be shown here.</p>
        </div>
      </section>
    </>
  )
}