import { FiAlertCircle, FiBookOpen, FiExternalLink, FiGithub } from 'react-icons/fi'
import '../../styles/DashboardPage.css'
import '../../styles/HelpPage.css'

export default function HelpPage() {
  return (
    <section className="help-page">
      <div className="help-hero">
        <div className="help-hero-icon">
          <FiGithub />
        </div>

        <div className="help-hero-content">
          <h1>Help & Support</h1>
          <p>
            If you encounter an issue while using KubeEvalHub, report it through the
            official GitHub repository. Include enough detail to help with investigation
            and resolution.
          </p>

          <a
            className="help-primary-button"
            href="https://github.com/Akaud/KubeEvalHub/issues"
            target="_blank"
            rel="noreferrer"
          >
            <FiExternalLink />
            <span>Open GitHub Issues</span>
          </a>
        </div>
      </div>

      <div className="help-grid">
        <article className="help-info-card">
          <div className="help-info-icon">
            <FiAlertCircle />
          </div>
          <div>
            <h3>Report a Technical Issue</h3>
            <p>
              Submit a GitHub issue for bugs, unexpected behavior, failed workflows, or
              other technical problems affecting the platform.
            </p>
          </div>
        </article>

        <article className="help-info-card">
          <div className="help-info-icon">
            <FiBookOpen />
          </div>
          <div>
            <h3>Include Relevant Details</h3>
            <p>
              Provide a concise summary, steps to reproduce the issue, expected behavior,
              actual behavior, and any logs or screenshots that may be useful.
            </p>
          </div>
        </article>
      </div>

      <div className="help-guidance-card">
        <h2>Recommended Before Submission</h2>

        <ul className="help-checklist">
          <li>Confirm the issue is reproducible</li>
          <li>Review existing GitHub issues to avoid duplicates</li>
          <li>Attach logs, error output, or screenshots when available</li>
          <li>Use a clear and specific issue title</li>
        </ul>

        <p className="help-guidance-note">
          GitHub issues are the preferred support channel for defect reporting and
          platform-related troubleshooting.
        </p>
      </div>
    </section>
  )
}