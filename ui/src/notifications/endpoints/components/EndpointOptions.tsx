// Libraries
import React, {FC, ChangeEvent} from 'react'

// Components
import EndpointOptionsSlack from './EndpointOptionsSlack'
import EndpointOptionsPagerDuty from './EndpointOptionsPagerDuty'
import EndpointOptionsHTTP from './EndpointOptionsHTTP'
import EndpointOptionsTelegram from './EndpointOptionsTelegram'
import EndpointOptionsTeams from './EndpointOptionsTeams'

// Types
import {
  NotificationEndpoint,
  SlackNotificationEndpoint,
  PagerDutyNotificationEndpoint,
  HTTPNotificationEndpoint,
  TelegramNotificationEndpoint,
  TeamsNotificationEndpoint,
} from 'src/types'

interface Props {
  endpoint: NotificationEndpoint
  onChange: (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void
  onChangeParameter: (key: string) => (value: string) => void
}

const EndpointOptions: FC<Props> = ({
  endpoint,
  onChange,
  onChangeParameter,
}) => {
  switch (endpoint.type) {
    case 'slack': {
      const {url} = endpoint as SlackNotificationEndpoint
      return <EndpointOptionsSlack url={url} onChange={onChange} />
    }
    case 'pagerduty': {
      const {clientURL, routingKey} = endpoint as PagerDutyNotificationEndpoint
      return (
        <EndpointOptionsPagerDuty
          clientURL={clientURL}
          routingKey={routingKey}
          onChange={onChange}
        />
      )
    }
    case 'telegram': {
      const {token, channel} = endpoint as TelegramNotificationEndpoint
      return (
        <EndpointOptionsTelegram
          token={token}
          channel={channel}
          onChange={onChange}
        />
      )
    }
    case 'http': {
      const {
        url,
        token,
        username,
        password,
        method,
        authMethod,
        contentTemplate,
      } = endpoint as HTTPNotificationEndpoint
      return (
        <EndpointOptionsHTTP
          onChange={onChange}
          onChangeParameter={onChangeParameter}
          url={url}
          token={token}
          username={username}
          password={password}
          method={method}
          authMethod={authMethod}
          contentTemplate={contentTemplate}
        />
      )
    }

    case 'teams': {
      const {url, secretURLSuffix} = endpoint as TeamsNotificationEndpoint
      return (
        <EndpointOptionsTeams
          url={url}
          secretURLSuffix={secretURLSuffix}
          onChange={onChange}
        />
      )
    }

    default:
      throw new Error(
        `Unknown endpoint type for endpoint: ${JSON.stringify(
          endpoint,
          null,
          2
        )}`
      )
  }
}

export default EndpointOptions