import rss from '@astrojs/rss';
import type { APIContext } from 'astro';
import appsData from '../../data/apps.json';

interface Release {
  version: string;
  date: string;
  title: string;
  description?: string;
  url?: string;
  type: string;
}

interface App {
  id: string;
  name: string;
  summary: string;
  icon?: string;
  currentReleaseVersion?: string;
  currentReleaseDate?: string;
  flathubUrl: string;
  isVerified?: boolean;
  packageType?: string;
  releases?: Release[];
}

export async function GET(context: APIContext) {
  const apps = (appsData.apps as App[]) || [];
  
  // Filter for Homebrew packages only
  const homebrewApps = apps.filter(app => app.packageType === 'homebrew');
  
  // Create a flat list of releases with their app context
  const allReleases = homebrewApps.flatMap(app => 
    (app.releases || []).map(release => ({
      app,
      release,
      parsedDate: new Date(release.date)
    }))
  );
  
  // Sort by date descending (newest first)
  allReleases.sort((a, b) => b.parsedDate.getTime() - a.parsedDate.getTime());
  
  // Take the latest 50 releases
  const recentReleases = allReleases.slice(0, 50);
  
  return rss({
    title: 'Bluefin Firehose - Homebrew Releases',
    description: 'Latest Homebrew package release updates',
    site: context.site || 'https://castrojo.github.io/bluefin-releases/',
    items: recentReleases.map(({ app, release, parsedDate }) => {
      const description = release.description 
        ? stripHtmlTags(release.description).substring(0, 500) 
        : app.summary;
      
      return {
        title: `${app.name} ${release.version}`,
        pubDate: parsedDate,
        description: description,
        link: app.flathubUrl,
        categories: ['homebrew'],
      };
    }),
    customData: `<language>en-us</language>`,
  });
}

function stripHtmlTags(html: string): string {
  return html
    .replace(/<[^>]*>/g, '')
    .replace(/\s+/g, ' ')
    .trim();
}
