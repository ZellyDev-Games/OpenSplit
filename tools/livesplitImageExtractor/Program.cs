// compile: mcs Program.cs -r:System.Drawing.dll -r:System.Xml.Linq.dll
// run: mono Program.exe <input>.lss <output>/

using System;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Linq;
using System.Runtime.Serialization.Formatters.Binary;
using System.Xml.Linq;

class Program
{
  static int Main(string[] args)
  {
    if (args.Length != 2)
    {
      Console.WriteLine("Usage: lssicons <input.lss> <output-dir>");
      return 1;
    }

    string input = args[0];
    string output = args[1];

    Directory.CreateDirectory(output);

    XDocument doc = XDocument.Load(input);

    var segments = doc.Descendants("Segment").ToList();

    int index = 0;

    foreach (var segment in segments)
    {
      XElement iconElement = segment.Element("Icon");

      if (iconElement == null || String.IsNullOrWhiteSpace(iconElement.Value))
      {
        index++;
        continue;
      }

      try
      {
        byte[] bytes = Convert.FromBase64String(iconElement.Value);

        using (MemoryStream ms = new MemoryStream(bytes))
        {
          Image image = (Image)new BinaryFormatter().Deserialize(ms);

          string filename = Path.Combine(output,
              index.ToString("D3") + ".png");

          image.Save(filename, ImageFormat.Png);

          image.Dispose();

          Console.WriteLine(filename);
        }
      }
      catch (Exception ex)
      {
        Console.Error.WriteLine(
            "Segment " + index + ": " + ex.Message);
      }

      index++;
    }

    return 0;
  }
}
