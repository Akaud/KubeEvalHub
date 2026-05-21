import {
  FiBookOpen,
  FiCheckCircle,
  FiCpu,
  FiLink,
} from 'react-icons/fi'
import { Link } from 'react-router-dom'
import '../../styles/DashboardPage.css'
import '../../styles/HelpPage.css'

export default function GuidePage() {
  return (
    <section className="help-page">
      <div className="help-hero">
        <div className="help-hero-icon">
          <FiBookOpen />
        </div>

        <div className="help-hero-content">
          <h1>Guide</h1>
          <p>
            Follow this guide to create a cluster, configure an agent, and connect
            your Kubernetes cluster to KubeEvalHub.
          </p>
        </div>
      </div>

      <div className="help-guide-layout">
        <div className="help-guide-left">
          <article className="help-info-card">
            <div className="help-info-icon">
              <FiCpu />
            </div>
            <div>
              <h3>Connection Workflow</h3>
              <p>
                Create a cluster, create an agent, assign the agent to the cluster,
                apply the generated YAML manifest, and wait for the cluster to come
                online.
              </p>
            </div>
          </article>

          <article className="help-info-card">
            <div className="help-info-icon">
              <FiLink />
            </div>
            <div>
              <h3>Required Kubernetes Access</h3>
              <p>
                Ensure that your local environment has kubectl access to the target
                Kubernetes cluster and permission to apply the generated manifest.
              </p>
            </div>
          </article>
        </div>

        <div className="help-guidance-card">
          <h2>Cluster Setup Steps</h2>

          <ul className="help-checklist">
            <li>
              <FiCheckCircle />
              <span>
                Open the <Link to="/dashboard/clusters">Clusters</Link> page and
                create a new cluster.
              </span>
            </li>

            <li>
              <FiCheckCircle />
              <span>
                Open the <Link to="/dashboard/agents">Agents</Link> page and
                create a new agent.
              </span>
            </li>

            <li>
              <FiCheckCircle />
              <span>
                Copy the generated YAML configuration and save it locally as a YAML
                file, for example <code>agent.yaml</code>.
              </span>
            </li>

            <li>
              <FiCheckCircle />
              <span>
                Return to the <Link to="/dashboard/clusters">Clusters</Link> page,
                open the cluster action menu by clicking the three-dot button in
                the upper-right corner of the cluster card, and select{' '}
                <strong>Assign Agent</strong>.
              </span>
            </li>

            <li>
              <FiCheckCircle />
              <span>
                In the assignment dialog, select the agent you created and confirm
                the assignment.
              </span>
            </li>

            <li>
              <FiCheckCircle />
              <span>
                Apply the saved YAML file to your Kubernetes cluster with{' '}
                <code>kubectl apply -f agent.yaml</code>.
              </span>
            </li>

            <li>
              <FiCheckCircle />
              <span>
                Wait for the cluster status to change to <strong>Online</strong>.
                This confirms that the agent has connected successfully.
              </span>
            </li>
          </ul>

          <p className="help-guidance-note">
            If the cluster remains offline, verify the agent assignment, confirm
            that the manifest was applied successfully, and check the agent pod logs
            in your Kubernetes cluster.
          </p>
        </div>
      </div>
    </section>
  )
}